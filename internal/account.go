package internal

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"strconv"
)

type AccountRepo struct {
	account     *Account
	storage     Storage
	distributor Messenger
	logger      *zap.Logger
}

type Account struct {
	Id           int            `json:"-"`
	Name         string         `json:"name"`
	Balance      float64        `json:"balance"`
	Transactions *[]Transaction `json:"transactions"`
}

type Transaction struct {
	OrderID       int     `json:"orderID"`
	OperationType string  `json:"operationType"`
	Amount        float64 `json:"amount"`
	Status        bool    `json:"status"`
}

type StorageRespond struct {
	Id      string
	Name    string
	Balance string
}

type IncStruct struct {
	ClientID   int
	ClientName string
	OrderID    int
	TypeTXN    string
	Amount     float64
}

func NewAccountRepo(storage Storage, distributor Messenger, logger *zap.Logger) *AccountRepo {
	return &AccountRepo{
		account:     new(Account),
		storage:     storage,
		distributor: distributor,
		logger:      logger,
	}
}

type AccountService interface {
	BalanceSufficient(ctx context.Context, incStruct *IncStruct) (bool, error)
}

type Storage interface {
	Select(ctx context.Context, accountID int) (*StorageRespond, error)
	Update(ctx context.Context, task chan IncStruct, finishedTask chan bool) error
}

type Messenger interface {
	SendMessage(ctx context.Context, domainStruct *IncStruct) error
	ReadMessage(ctx context.Context, bindingKey string, finishedTask chan bool) (chan IncStruct, error)
}

func (a *AccountRepo) BalanceSufficient(ctx context.Context, incStruct *IncStruct) (bool, error) {
	dbRespond, err := a.storage.Select(ctx, incStruct.ClientID)
	if err != nil {
		return false, fmt.Errorf("getting balance from db %v", zap.Error(err))
	}
	if dbRespond == nil {
		return false, fmt.Errorf("client is not in the db %v", zap.Error(err))
	}

	if incStruct.TypeTXN == "deposit" {
		err = a.putMessageInQueue(ctx, incStruct)
		if err != nil {
			return false, fmt.Errorf("performing transaction %v", zap.Error(err))
		}
		go a.updateAccount(ctx, incStruct)
		return true, nil
	}

	accountBalance, err := strconv.ParseFloat(dbRespond.Balance, 64)
	if err != nil {
		return false, fmt.Errorf("converting string to float64 %v", zap.Error(err))
	}
	newBalance := accountBalance - incStruct.Amount
	if newBalance < 0 {
		return false, nil
	}

	err = a.putMessageInQueue(ctx, incStruct)
	if err != nil {
		return false, fmt.Errorf("performing transaction %v", zap.Error(err))
	}
	go a.updateAccount(ctx, incStruct)

	return true, nil
}

func (a *AccountRepo) putMessageInQueue(ctx context.Context, incStruct *IncStruct) error {
	err := a.distributor.SendMessage(ctx, incStruct)
	if err != nil {
		return fmt.Errorf("putting new message in a queue %v", zap.Error(err))
	}
	return nil
}

func (a *AccountRepo) updateAccount(ctx context.Context, incStruct *IncStruct) (bool, error) {
	finishedTask := make(chan bool)
	deliveredMsgs, err := a.distributor.ReadMessage(ctx, strconv.Itoa(incStruct.ClientID), finishedTask)
	if err != nil {
		return false, fmt.Errorf("performing transaction %v", zap.Error(err))
	}
	err = a.storage.Update(ctx, deliveredMsgs, finishedTask)
	if err != nil {
		return false, fmt.Errorf("updating client balance %v", zap.Error(err))
	}
	return true, nil
}
