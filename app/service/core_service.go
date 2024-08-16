package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/atrariksa/kenalan-core/app/model"
	"github.com/atrariksa/kenalan-core/app/repository"
	"github.com/atrariksa/kenalan-core/app/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/atrariksa/kenalan-core/app/external/grpc_client"
)

var KeyViewProfile = "view_profile:%s"

type ICoreService interface {
	SignUp(ctx context.Context, signUpRequest model.SignUpRequest) error
	Login(ctx context.Context, loginRequest model.LoginRequest) (string, error)
	ViewProfile(ctx context.Context, vpRequest model.ViewProfileRequest) (model.Profile, error)
}

type CoreService struct {
	Repo      repository.ICoreRepository
	RedisRepo repository.IRedisCoreRepository
}

func NewCoreService(coreRepo repository.ICoreRepository, redisRepo repository.IRedisCoreRepository) *CoreService {
	return &CoreService{
		Repo:      coreRepo,
		RedisRepo: redisRepo,
	}
}

func (cs *CoreService) SignUp(ctx context.Context, signUpRequest model.SignUpRequest) error {
	conn, err := grpc.NewClient("localhost:6021", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("did not connect: %v", err)
		return errors.New("internal error")
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	r, err := c.IsUserExist(gCtx, &pb.IsUserExistRequest{Email: signUpRequest.Email})
	if err != nil {
		log.Printf("call IsUserExist failed: %v", err)
		return errors.New("internal error")
	}
	log.Printf("IsUserExist: %v", r.IsUserExist)

	if r.IsUserExist {
		return errors.New("user already exists")
	}

	gCtx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rUser, err := c.CreateUser(gCtx, &pb.CreateUserRequest{
		User: &pb.User{
			FullName: signUpRequest.Fullname,
			Gender:   signUpRequest.Gender,
			Dob:      signUpRequest.DOB,
			Email:    signUpRequest.Email,
			Password: signUpRequest.Password,
		},
	})
	if err != nil {
		log.Printf("call CreateUser failed: %v", err)
		return errors.New("internal error")
	}
	log.Printf("CreateUser: %v", rUser.Message)

	return nil
}

func (cs *CoreService) Login(ctx context.Context, loginRequest model.LoginRequest) (string, error) {
	user, err := HandleGetUserByEmail(ctx, loginRequest)
	if err != nil {
		return "", errors.New("invalid email or password 1")
	}

	err = util.ValidatePassword(loginRequest.Password, user.User.Password)
	if err != nil {
		return "", errors.New("invalid email or password 2")
	}

	rToken, err := HandleGetToken(ctx, loginRequest)
	if err != nil {
		return "", errors.New("invalid email or password 3")
	}

	return rToken.Token, nil
}

func (cs *CoreService) ViewProfile(ctx context.Context, vpRequest model.ViewProfileRequest) (model.Profile, error) {
	rToken, err := HandleIsTokenValid(ctx, vpRequest)
	if err != nil {
		return model.Profile{}, err
	}

	viewProfileData, err := cs.RedisRepo.GetViewProfile(ctx, fmt.Sprintf(KeyViewProfile, rToken.Email))
	if err != nil {
		return model.Profile{}, err
	}

	if viewProfileData.Email == "" {
	}

	return model.Profile{}, nil
}

var HandleGetUserByEmail = func(ctx context.Context, loginRequest model.LoginRequest) (*pb.GetUserByEmailResponse, error) {
	conn, err := grpc.NewClient("localhost:6021", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New("internal error")
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rUser, err := c.GetUserByEmail(gCtx, &pb.GetUserByEmailRequest{Email: loginRequest.Email})
	if err != nil {
		log.Printf("call GetUserByEmail failed: %v", err)
		return nil, errors.New("internal error")
	}
	log.Printf("GetUserByEmail: %v", rUser.User.Id)

	if rUser.User.Id == 0 {
		return nil, errors.New("user not found")
	}

	return rUser, nil
}

var HandleGetToken = func(ctx context.Context, loginRequest model.LoginRequest) (*pb.GetTokenResponse, error) {
	conn, err := grpc.NewClient("localhost:6022", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New("internal error")
	}
	defer conn.Close()
	c := pb.NewAuthServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rToken, err := c.GetToken(gCtx, &pb.GetTokenRequest{Email: loginRequest.Email})
	if err != nil {
		log.Printf("call GetToken failed: %v", err)
		return nil, errors.New("internal error")
	}
	log.Printf("GetToken: %v", rToken.Code)

	if rToken.Token == "" {
		return nil, errors.New("internal error")
	}

	return rToken, nil
}

var HandleIsTokenValid = func(ctx context.Context, vpRequest model.ViewProfileRequest) (*pb.IsTokenValidResponse, error) {
	conn, err := grpc.NewClient("localhost:6022", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New("internal error")
	}
	defer conn.Close()
	c := pb.NewAuthServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rToken, err := c.IsTokenValid(gCtx, &pb.IsTokenValidRequest{Token: vpRequest.Token})
	if err != nil {
		log.Printf("call IsTokenValid failed: %v", err)
		return nil, errors.New("internal error")
	}
	log.Printf("IsTokenValid: %v", rToken.Code)

	if !rToken.IsTokenValid {
		return nil, errors.New("unauthorized")
	}

	return rToken, nil
}
