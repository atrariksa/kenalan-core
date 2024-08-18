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
	"github.com/atrariksa/kenalan-core/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/atrariksa/kenalan-core/app/external/grpc_client"
)

var KeyViewProfile = "view_profile:%s"

type ICoreService interface {
	SignUp(ctx context.Context, signUpRequest model.SignUpRequest) error
	Login(ctx context.Context, loginRequest model.LoginRequest) (string, error)
	ViewProfile(ctx context.Context, vpRequest model.ViewProfileRequest) (model.Profile, error)
	Purchase(ctx context.Context, pr model.PurchaseRequest) error
}

type CoreService struct {
	Repo      repository.ICoreRepository
	RedisRepo repository.IRedisCoreRepository
	Cfg       *config.Config
}

func NewCoreService(
	coreRepo repository.ICoreRepository,
	redisRepo repository.IRedisCoreRepository,
	cfg *config.Config) *CoreService {

	return &CoreService{
		Repo:      coreRepo,
		RedisRepo: redisRepo,
		Cfg:       cfg,
	}
}

func (cs *CoreService) SignUp(ctx context.Context, signUpRequest model.SignUpRequest) error {
	conn, err := GetUserServiceConnection(cs.Cfg.UserServerConfig.Host, cs.Cfg.UserServerConfig.Port)
	if err != nil {
		log.Printf("did not connect: %v", err)
		return errors.New(util.ErrInternalError)
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	r, err := c.IsUserExist(gCtx, &pb.IsUserExistRequest{Email: signUpRequest.Email})
	if err != nil {
		log.Printf("call IsUserExist failed: %v", err)
		return errors.New(util.ErrInternalError)
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
		return errors.New(util.ErrInternalError)
	}
	log.Printf("CreateUser: %v", rUser.Message)

	return nil
}

func (cs *CoreService) Login(ctx context.Context, loginRequest model.LoginRequest) (string, error) {
	user, err := HandleGetUserByEmail(ctx, cs.Cfg, loginRequest)
	if err != nil {
		return "", errors.New("invalid email or password 1")
	}

	err = util.ValidatePassword(loginRequest.Password, user.User.Password)
	if err != nil {
		return "", errors.New("invalid email or password 2")
	}

	rToken, err := HandleGetToken(ctx, cs.Cfg, loginRequest)
	if err != nil {
		return "", errors.New("invalid email or password 3")
	}

	return rToken.Token, nil
}

func (cs *CoreService) ViewProfile(ctx context.Context, vpRequest model.ViewProfileRequest) (model.Profile, error) {
	var nextProfile model.Profile
	rToken, err := HandleIsTokenValid(ctx, cs.Cfg, vpRequest.Token)
	if err != nil {
		return nextProfile, err
	}

	if rToken.Email == "" {
		return nextProfile, errors.New(util.ErrInvalidToken)
	}

	viewProfileData, err := cs.RedisRepo.GetViewProfile(ctx, fmt.Sprintf(KeyViewProfile, rToken.Email))
	if err != nil {
		return nextProfile, err
	}

	var rUser *pb.GetUserSubscriptionResponse
	if viewProfileData.Email == "" {
		rUser, err = HandleGetUserSubscription(ctx, cs.Cfg, vpRequest, viewProfileData.Email)
		if err != nil {
			return nextProfile, errors.New(util.ErrInternalError)
		}

		for i := 0; i < len(rUser.Subscriptions); i++ {
			if rUser.Subscriptions[i].ProductCode == util.UnlimitedSwipeProductCode {
				viewProfileData.IsUnlimitedSwipe = true

				// TODO: Handle for subscription expired_at less than 24 hour
				// - add delayed job for worker to update value IsUnlimitedSwipe to false

				break
			}
		}

		viewProfileData.ViewerID = rUser.User.Id
		viewProfileData.Email = rToken.Email
		viewProfileData.ViewedProfileIDs = make([]int64, 0)
		viewProfileData.ViewerGender = rUser.User.Gender
		err = cs.RedisRepo.StoreViewProfile(ctx, fmt.Sprintf(KeyViewProfile, rToken.Email), viewProfileData)
		if err != nil {
			return nextProfile, errors.New(util.ErrInternalError)
		}
	}

	// handle swipe count
	if viewProfileData.SwipeCount >= 10 && !viewProfileData.IsUnlimitedSwipe {
		return nextProfile, errors.New("already used up all swipe quota")
	}

	if vpRequest.SwipeLeft {
		// pass: get next profile
		excludeIDs := viewProfileData.ViewedProfileIDs
		log.Println(viewProfileData.ViewedProfileIDs)
		nextProfileGender := "F"
		if viewProfileData.ViewerGender == "F" {
			nextProfileGender = "M"
		}
		excludeIDs = append(excludeIDs, viewProfileData.ViewerID)

		rNextProfile, err := HandleGetNextProfileExceptIDs(ctx, cs.Cfg, excludeIDs, nextProfileGender)
		if err != nil {
			return nextProfile, err
		}

		viewProfileData.ViewedProfileIDs = append(viewProfileData.ViewedProfileIDs, rNextProfile.User.Id)
		viewProfileData.SwipeCount++

		err = cs.RedisRepo.StoreViewProfile(ctx, fmt.Sprintf(KeyViewProfile, rToken.Email), viewProfileData)
		if err != nil {
			return nextProfile, errors.New(util.ErrInternalError)
		}

		nextProfile.ID = rNextProfile.User.Id
		nextProfile.Fullname = rNextProfile.User.FullName
		nextProfile.PhotoURL = rNextProfile.User.PhotoUrl
		for i := 0; i < len(rNextProfile.Subscriptions); i++ {
			if rNextProfile.Subscriptions[i].ProductCode == util.AccountVerifiedProductCode {
				nextProfile.IsVerified = true
				break
			}
		}
	} else {
		// like:
		viewProfileData.SwipeCount++
		err = cs.RedisRepo.StoreViewProfile(ctx, fmt.Sprintf(KeyViewProfile, rToken.Email), viewProfileData)
		if err != nil {
			return nextProfile, errors.New(util.ErrInternalError)
		}
	}

	return nextProfile, nil
}

func (cs *CoreService) Purchase(ctx context.Context, pr model.PurchaseRequest) error {
	rToken, err := HandleIsTokenValid(ctx, cs.Cfg, pr.Token)
	if err != nil {
		return err
	}

	if rToken.Email == "" {
		return errors.New(util.ErrInvalidToken)
	}

	_, err = HandleUpsertSubscription(ctx, cs.Cfg, pr, rToken.Email)
	if err != nil {
		return err
	}

	viewProfileData, _ := cs.RedisRepo.GetViewProfile(ctx, fmt.Sprintf(KeyViewProfile, rToken.Email))
	if viewProfileData.Email == rToken.Email {
		if pr.ProductCode == util.UnlimitedSwipeProductCode {
			viewProfileData.IsUnlimitedSwipe = true
			cs.RedisRepo.StoreViewProfile(ctx, fmt.Sprintf(KeyViewProfile, rToken.Email), viewProfileData)
		}
	}

	return nil
}

var HandleGetUserSubscription = func(
	ctx context.Context,
	cfg *config.Config,
	viewProfileRequest model.ViewProfileRequest,
	email string) (*pb.GetUserSubscriptionResponse, error) {

	conn, err := GetUserServiceConnection(cfg.UserServerConfig.Host, cfg.UserServerConfig.Port)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rUser, err := c.GetUserSubscription(gCtx, &pb.GetUserSubscriptionRequest{
		Email: email,
	})
	if err != nil {
		log.Printf("call GetUserSubscription failed: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}

	if rUser.User.Id == 0 {
		return nil, errors.New("user not found")
	}

	return rUser, nil
}

var HandleGetNextProfileExceptIDs = func(
	ctx context.Context,
	cfg *config.Config,
	ids []int64,
	gender string) (*pb.GetNextProfileExceptIDsResponse, error) {

	conn, err := GetUserServiceConnection(cfg.UserServerConfig.Host, cfg.UserServerConfig.Port)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rUser, err := c.GetNextProfileExceptIDs(gCtx, &pb.GetNextProfileExceptIDsRequest{
		Ids:    ids,
		Gender: gender,
	})

	if err != nil {
		if status.Code(err) == 05 {
			return nil, errors.New("user not found")
		}
		log.Printf("call GetNextProfileExceptIDs failed: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}

	if rUser.User.Id == 0 {
		return nil, errors.New("user not found")
	}

	return rUser, nil
}

var HandleGetUserByEmail = func(
	ctx context.Context,
	cfg *config.Config,
	loginRequest model.LoginRequest) (*pb.GetUserByEmailResponse, error) {

	conn, err := GetUserServiceConnection(cfg.UserServerConfig.Host, cfg.UserServerConfig.Port)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rUser, err := c.GetUserByEmail(gCtx, &pb.GetUserByEmailRequest{Email: loginRequest.Email})
	if err != nil {
		log.Printf("call GetUserByEmail failed: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}

	if rUser.User.Id == 0 {
		return nil, errors.New("user not found")
	}

	return rUser, nil
}

var HandleGetToken = func(
	ctx context.Context,
	cfg *config.Config,
	loginRequest model.LoginRequest) (*pb.GetTokenResponse, error) {

	conn, err := GetAuthServiceConnection(cfg.AuthServerConfig.Host, cfg.AuthServerConfig.Port)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}
	defer conn.Close()
	c := pb.NewAuthServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rToken, err := c.GetToken(gCtx, &pb.GetTokenRequest{Email: loginRequest.Email})
	if err != nil {
		log.Printf("call GetToken failed: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}

	if rToken.Token == "" {
		return nil, errors.New(util.ErrInternalError)
	}

	return rToken, nil
}

var HandleIsTokenValid = func(
	ctx context.Context,
	cfg *config.Config,
	token string) (*pb.IsTokenValidResponse, error) {

	conn, err := GetAuthServiceConnection(cfg.AuthServerConfig.Host, cfg.AuthServerConfig.Port)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}
	defer conn.Close()
	c := pb.NewAuthServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rToken, err := c.IsTokenValid(gCtx, &pb.IsTokenValidRequest{Token: token})
	if err != nil {
		if status.Code(err) == util.CodeInvalidToken {
			return nil, errors.New(util.ErrUnauthorized)
		}
		log.Printf("call IsTokenValid failed: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}

	if !rToken.IsTokenValid {
		return nil, errors.New(util.ErrUnauthorized)
	}

	return rToken, nil
}

var HandleUpsertSubscription = func(
	ctx context.Context,
	cfg *config.Config,
	purchaseRequest model.PurchaseRequest,
	email string) (*pb.UpsertSubscriptionResponse, error) {

	conn, err := GetUserServiceConnection(cfg.UserServerConfig.Host, cfg.UserServerConfig.Port)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, errors.New(util.ErrInternalError)
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)

	gCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rUpsertSubscription, err := c.UpsertSubscription(gCtx, &pb.UpsertSubscriptionRequest{
		UserId:      purchaseRequest.UserID,
		Email:       email,
		ProductCode: purchaseRequest.ProductCode,
		ProductName: purchaseRequest.ProductName,
		ExpiredAt:   purchaseRequest.ExpiredAt,
	})
	if err != nil {
		if status.Code(err) == 14 {
			log.Printf("call UpsertSubscription failed: %v", err)
			return nil, errors.New(util.ErrInternalError)
		}
		return nil, errors.New(util.ErrProductNotFound)
	}

	return rUpsertSubscription, nil
}

var GetUserServiceConnection = func(host string, port int) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		fmt.Sprintf("%v:%v", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

var GetAuthServiceConnection = func(host string, port int) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		fmt.Sprintf("%v:%v", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}
