package postgres

import (
	"context"
	"fmt"
	"log"
	"test/internal/config"
	"test/internal/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// структура бд
type Pgdb struct {
	Db *pgxpool.Pool
}

func CreateDB() (*Pgdb, error) {
	// создаем контекст
	ctx := context.Background()
	// парсим конфиг
	cfg, err := config.InitConfig()
	if err != nil {
		return nil, err
	}

	connectString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Db)
	// создаем pgxpool
	pool, err := pgxpool.New(ctx, connectString)
	if err != nil {
		return nil, err
	}
	return &Pgdb{
		Db: pool,
	}, nil

}

func (p *Pgdb) WriteData(order entities.Order) error {
	tag, err := p.Db.Exec(context.Background(),
		`INSERT INTO orders 
    (order_uid,
    track_number,
    entry,
    delivery,
    payment,
    items,
    locale_chr,
    internal_signature,
    customer_id,
    delivery_service,
    shardkey,
    sm_id,
    data_created,
    oof_shard) VALUES (
    @order_uid,
    @track_number,
    @entry,
    @delivery,
    @payment,
    @items,
    @locale_chr,
    @internal_signature,
    @customer_id,
    @delivery_service,
    @shardkey,
    @sm_id,
    @data_created,
    @oof_shard)`, pgx.NamedArgs{
			"order_uid":          order.OrderUID,
			"track_number":       order.TrackNumber,
			"entry":              order.Entry,
			"delivery":           order.DeliveryString(),
			"payment":            order.PaymentString(),
			"items":              order.ItemsString(),
			"locale_chr":         order.Locale,
			"internal_signature": order.InternalSignature,
			"customer_id":        order.CustomerID,
			"delivery_service":   order.DeliveryService,
			"shardkey":           order.Shardkey,
			"sm_id":              order.SmID,
			"data_created":       order.DateCreated,
			"oof_shard":          order.OofShard,
		})

	if err != nil {
		return err
	}
	r := tag.RowsAffected()
	log.Printf("Postgres data written. Rows affected %d \n", r)
	return nil
}

func (pg *Pgdb) RecoverData() (pgx.Rows, error) {
	var ctx = context.Background()
	rows, err := pg.Db.Query(ctx,
		`SELECT 
order_uid,
track_number,
entry,
json_build_object(
    'name', (delivery).name,
    'phone', (delivery).phone,
    'zip', (delivery).zip,
    'city', (delivery).city,
    'address', (delivery).address,
    'region', (delivery).region,
    'email', (delivery).email
) AS delivery,
json_build_object(
	'transaction', (payment).transaction,
	'request_id', (payment).request_id,
	'currency',(payment).currency,
	'provider',(payment).provider,
	'amount',(payment).amount,
	'payment_dt',(payment).payment_dt,
	'bank',(payment).bank,
	'delivery_cost',(payment).delivery_cost,
	'goods_total',(payment).goods_total,
	'custom_fee',(payment).custom_fee
) AS payment,
json_build_object('items',items) AS items,
locale_chr,
internal_signature,
customer_id,
delivery_service,
shardkey,
sm_id,
data_created,
oof_shard 
FROM orders`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
