package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameDeviceInfo = "device_info"

func GetDeviceInfoList(ctx context.Context, cond *model.DeviceInfo, option QueryOption) ([]*model.DeviceInfo, int64, error) {
	var args []interface{}
	where := `WHERE device_id <> '' AND active_status = 1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if cond.UserID != "" {
		where += ` AND user_id = ?`
		args = append(args, cond.UserID)
	}
	if cond.DeviceStatus != "" && cond.DeviceStatus != "allDevices" {
		where += ` AND device_status = ?`
		args = append(args, cond.DeviceStatus)
	}
	if cond.IpLocation != "" && cond.IpLocation != "0" {
		where += ` AND ip_location = ?`
		args = append(args, cond.IpLocation)
	}
	if cond.NodeType > 0 {
		where += ` AND node_type = ?`
		args = append(args, cond.NodeType)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.DeviceInfo

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameDeviceInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s ORDER BY device_rank LIMIT %d OFFSET %d`, tableNameDeviceInfo, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func GetDeviceInfoByID(ctx context.Context, deviceID string) (*model.DeviceInfo, error) {
	var out model.DeviceInfo
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id = ?`, tableNameDeviceInfo), deviceID,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func UpdateUserDeviceInfo(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET user_id = :user_id, updated_at = now(),bind_status = :bind_status WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}
func UpdateDeviceName(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET updated_at = now(),device_name = :device_name WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}
func BulkUpsertDeviceInfo(ctx context.Context, deviceInfos []*model.DeviceInfo) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statement := upsertDeviceInfoStatement()
	for _, deviceInfo := range deviceInfos {
		_, err = tx.NamedExecContext(ctx, statement, deviceInfo)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func BulkUpdateDeviceInfo(ctx context.Context, deviceInfos []*model.DeviceInfo) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, deviceInfo := range deviceInfos {
		_, err = tx.NamedExecContext(ctx, fmt.Sprintf(
			`UPDATE %s SET today_online_time = :today_online_time, today_profit = :today_profit,
				yesterday_profit = :yesterday_profit, seven_days_profit = :seven_days_profit, month_profit = :month_profit, 
				updated_at = now() WHERE device_id = :device_id`, tableNameDeviceInfo),
			deviceInfo)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func upsertDeviceInfoStatement() string {
	insertStatement := fmt.Sprintf(
		`INSERT INTO %s (device_id, node_type, device_name, user_id, sn_code, operator,
				network_type, system_version, product_type, active_status,
				network_info, external_ip, internal_ip, ip_location, ip_country, ip_city, mac_location, nat_type, upnp,
				pkg_loss_ratio, nat_ratio, latency, cpu_usage, memory_usage, cpu_cores, memory, disk_usage, disk_space, work_status,
				device_status, disk_type, io_system, online_time, today_online_time, today_profit, total_upload, total_download, download_count, block_count,
				yesterday_profit, seven_days_profit, month_profit, cumulative_profit, bandwidth_up, bandwidth_down, created_at, updated_at)
			VALUES (:device_id, :node_type, :device_name, :user_id, :sn_code, :operator,
			    :network_type, :system_version, :product_type, :active_status,
			    :network_info, :external_ip, :internal_ip, :ip_location, :ip_country, :ip_city, :mac_location, :nat_type, :upnp, 
			    :pkg_loss_ratio, :nat_ratio, :latency, :cpu_usage, :memory_usage, :cpu_cores, :memory, :disk_usage, :disk_space, :work_status, 
			    :device_status, :disk_type, :io_system, :online_time, :today_online_time, :today_profit, :total_upload, :total_download, :download_count, block_count,
				:yesterday_profit, :seven_days_profit, :month_profit, :cumulative_profit, :bandwidth_up, :bandwidth_down, now(), now())`, tableNameDeviceInfo,
	)
	updateStatement := ` ON DUPLICATE KEY UPDATE node_type = :node_type,  device_name = :device_name,
				sn_code = :sn_code,  operator = :operator, network_type = :network_type, active_status = :active_status,
				system_version = :system_version,  product_type = :product_type, network_info = :network_info, cumulative_profit = :cumulative_profit,
				external_ip = :external_ip,  internal_ip = :internal_ip,  ip_location = :ip_location, ip_country = :ip_country, ip_city = :ip_city, 
				mac_location = :mac_location,  nat_type = :nat_type,  upnp = :upnp, pkg_loss_ratio = :pkg_loss_ratio, online_time = :online_time,
				nat_ratio = :nat_ratio,  latency = :latency,  cpu_usage = :cpu_usage, cpu_cores = :cpu_cores,  memory_usage = :memory_usage, memory = :memory,
				disk_usage = :disk_usage, disk_space = :disk_space,  work_status = :work_status, device_status = :device_status,  disk_type = :disk_type,
 				total_upload = :total_upload, total_download = :total_download, download_count = :download_count, block_count = :block_count,
				io_system = :io_system, bandwidth_up = :bandwidth_up, bandwidth_down = :bandwidth_down, updated_at = now()`
	return insertStatement + updateStatement
}

func GetAllAreaFromDeviceInfo(ctx context.Context) ([]string, error) {
	queryStatement := fmt.Sprintf(`SELECT ip_location FROM %s GROUP BY ip_location;`, tableNameDeviceInfo)
	var out []string
	err := DB.SelectContext(ctx, &out, queryStatement)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func SumFullNodeInfoFromDeviceInfo(ctx context.Context) (*model.FullNodeInfo, error) {
	queryStatement := fmt.Sprintf(`SELECT count( device_id ) AS total_node_count ,  SUM(IF(node_type = 1, 1, 0)) AS edge_count, 
       SUM(IF(node_type = 2, 1, 0)) AS candidate_count, SUM(IF(node_type = 3, 1, 0)) AS validator_count, ROUND(SUM( disk_space),4) AS total_storage, 
       ROUND(SUM(bandwidth_up),2) AS total_upstream_bandwidth, ROUND(SUM(bandwidth_down),2) AS total_downstream_bandwidth FROM %s where active_status = 1;`, tableNameDeviceInfo)

	var out model.FullNodeInfo
	if err := DB.QueryRowxContext(ctx, queryStatement).StructScan(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

type UserDeviceProfile struct {
	CumulativeProfit sql.NullFloat64 `json:"cumulative_profit" db:"cumulative_profit"`
	YesterdayProfit  sql.NullFloat64 `json:"yesterday_profit" db:"yesterday_profit"`
	TodayProfit      sql.NullFloat64 `json:"today_profit" db:"today_profit"`
	SevenDaysProfit  sql.NullFloat64 `json:"seven_days_profit" db:"seven_days_profit"`
	MonthProfit      sql.NullFloat64 `json:"month_profit" db:"month_profit"`
	TotalNum         sql.NullInt64   `json:"total_num" db:"total_num"`
	OnlineNum        sql.NullInt64   `json:"online_num" db:"online_num"`
	OfflineNum       sql.NullInt64   `json:"offline_num" db:"offline_num"`
	AbnormalNum      sql.NullInt64   `json:"abnormal_num" db:"abnormal_num"`
	TotalBandwidth   sql.NullFloat64 `json:"total_bandwidth" db:"total_bandwidth"`
}

func CountUserDeviceInfo(ctx context.Context, userID string) (*UserDeviceProfile, error) {
	queryStatement := fmt.Sprintf(`SELECT sum(cumulative_profit) as cumulative_profit, sum(yesterday_profit) as yesterday_profit, 
sum(today_profit) as today_profit, sum(seven_days_profit) as seven_days_profit, sum(month_profit) as month_profit, count(*) as total_num, 
count(IF(device_status = 'online', 1, NULL)) as online_num ,count(IF(device_status = 'offline', 1, NULL)) as offline_num, 
count(IF(device_status = 'abnormal', 1, NULL)) as abnormal_num, sum(bandwidth_up) as total_bandwidth  from %s where user_id = ? and active_status = 1;`, tableNameDeviceInfo)

	var out UserDeviceProfile
	if err := DB.QueryRowxContext(ctx, queryStatement, userID).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
