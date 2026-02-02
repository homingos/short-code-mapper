package transaction

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionManager interface {
	BeginTx(ctx context.Context) (mongo.SessionContext, error)
}

type transactionManager struct {
	client *mongo.Client
}

func NewTransactionManager(client *mongo.Client) TransactionManager {
	return &transactionManager{
		client: client,
	}
}

func (tm *transactionManager) BeginTx(ctx context.Context) (mongo.SessionContext, error) {
	session, err := tm.client.StartSession()
	if err != nil {
		return nil, err
	}

	if err = session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, err
	}

	return mongo.NewSessionContext(ctx, session), nil
}
