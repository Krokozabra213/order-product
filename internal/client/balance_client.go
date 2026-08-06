package client

import (
	"context"
	"fmt"

	balancev1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/balance/v1"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BalanceClient struct {
	api  balancev1.BalanceService_APIClient
	conn *grpc.ClientConn
}

func NewBalance(ctx context.Context, addr string) (*BalanceClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("balance client: grpc new client: %w", err)
	}

	return &BalanceClient{
		api:  balancev1.NewBalanceService_APIClient(conn),
		conn: conn,
	}, nil
}

func (c *BalanceClient) Close() error {
	return c.conn.Close()
}

func (c *BalanceClient) Charge(ctx context.Context, operationID string, userID, orderID uuid.UUID, amount int64) (*domain.BalanceStatus, error) {
	resp, err := c.api.OrderChange(ctx, &balancev1.OrderChangeRequest{
		OperationId: operationID,
		UserId:      userID.String(),
		OrderId:     orderID.String(),
		Amount:      int32(amount),
		Type:        balancev1.OrderOpType_ORDER_OP_TYPE_WITHDRAW,
	})
	if err != nil {
		return nil, fmt.Errorf("balance client: get account: %w", err)
	}

	var status domain.BalanceStatus
	switch resp.Status {
	case balancev1.OrderOpStatus_ORDER_OP_STATUS_SUCCESS:
		status = domain.BalanceStatusSuccess
	case balancev1.OrderOpStatus_ORDER_OP_STATUS_FAILED:
		status = domain.BalanceStatusFailed
	case balancev1.OrderOpStatus_ORDER_OP_STATUS_NOMONEY:
		status = domain.BalanceStatusNoMoney
	default:
		status = domain.BalanceStatusFailed
	}
	return &status, nil
}
