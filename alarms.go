package firewalla

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/url"
	"strconv"
)

// AlarmType is the documented enum 1..16.
type AlarmType int

const (
	AlarmSecurityActivity AlarmType = 1
	AlarmAbnormalUpload   AlarmType = 2
	AlarmLargeBandwidth   AlarmType = 3
	AlarmMonthlyDataPlan  AlarmType = 4
	AlarmNewDevice        AlarmType = 5
	AlarmDeviceBackOnline AlarmType = 6
	AlarmDeviceOffline    AlarmType = 7
	AlarmVideo            AlarmType = 8
	AlarmGaming           AlarmType = 9
	AlarmPorn             AlarmType = 10
	AlarmVPNActivity      AlarmType = 11
	AlarmVPNRestored      AlarmType = 12
	AlarmVPNError         AlarmType = 13
	AlarmOpenPort         AlarmType = 14
	AlarmInternetConn     AlarmType = 15
	AlarmLargeUpload      AlarmType = 16
)

func (t AlarmType) String() string {
	switch t {
	case AlarmSecurityActivity:
		return "SecurityActivity"
	case AlarmAbnormalUpload:
		return "AbnormalUpload"
	case AlarmLargeBandwidth:
		return "LargeBandwidth"
	case AlarmMonthlyDataPlan:
		return "MonthlyDataPlan"
	case AlarmNewDevice:
		return "NewDevice"
	case AlarmDeviceBackOnline:
		return "DeviceBackOnline"
	case AlarmDeviceOffline:
		return "DeviceOffline"
	case AlarmVideo:
		return "Video"
	case AlarmGaming:
		return "Gaming"
	case AlarmPorn:
		return "Porn"
	case AlarmVPNActivity:
		return "VPNActivity"
	case AlarmVPNRestored:
		return "VPNRestored"
	case AlarmVPNError:
		return "VPNError"
	case AlarmOpenPort:
		return "OpenPort"
	case AlarmInternetConn:
		return "InternetConnectivity"
	case AlarmLargeUpload:
		return "LargeUpload"
	default:
		return fmt.Sprintf("Unknown(%d)", int(t))
	}
}

// Alarm represents one alarm record.
type Alarm struct {
	AID      string         `json:"aid"`
	GID      string         `json:"gid"`
	Type     AlarmType      `json:"type"`
	TS       int64          `json:"ts"`
	Status   string         `json:"status"`
	Message  string         `json:"message,omitempty"`
	Device   *AlarmDevice   `json:"device,omitempty"`
	Remote   *AlarmRemote   `json:"remote,omitempty"`
	Transfer *AlarmTransfer `json:"transfer,omitempty"`
}

type AlarmDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type AlarmRemote struct {
	Domain   string `json:"domain"`
	Region   string `json:"region"`
	Category string `json:"category"`
}
type AlarmTransfer struct {
	Download int64 `json:"download"`
	Upload   int64 `json:"upload"`
	Total    int64 `json:"total"`
}

type AlarmListOptions struct {
	Query   string
	GroupBy string
	SortBy  string
	Limit   int
	Cursor  string
}

func alarmListQuery(opts *AlarmListOptions) url.Values {
	q := url.Values{}
	if opts == nil {
		return q
	}
	if opts.Query != "" {
		q.Set("query", opts.Query)
	}
	if opts.GroupBy != "" {
		q.Set("groupBy", opts.GroupBy)
	}
	if opts.SortBy != "" {
		q.Set("sortBy", opts.SortBy)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	return q
}

func (s *AlarmsService) List(ctx context.Context, opts *AlarmListOptions) (*Page[Alarm], error) {
	var out Page[Alarm]
	if err := s.client.do(ctx, "GET", "/alarms", alarmListQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AlarmsService) All(ctx context.Context, opts *AlarmListOptions) iter.Seq2[Alarm, error] {
	return paginate(ctx, func(cursor string) (*Page[Alarm], error) {
		o := AlarmListOptions{}
		if opts != nil {
			o = *opts
		}
		o.Cursor = cursor
		return s.List(ctx, &o)
	})
}

func (s *AlarmsService) Get(ctx context.Context, gid, aid string) (*Alarm, error) {
	if gid == "" || aid == "" {
		return nil, errors.New("firewalla: gid and aid are required")
	}
	var out Alarm
	path := "/alarms/" + url.PathEscape(gid) + "/" + url.PathEscape(aid)
	if err := s.client.do(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AlarmsService) Delete(ctx context.Context, gid, aid string) error {
	if gid == "" || aid == "" {
		return errors.New("firewalla: gid and aid are required")
	}
	path := "/alarms/" + url.PathEscape(gid) + "/" + url.PathEscape(aid)
	return s.client.do(ctx, "DELETE", path, nil, nil, nil)
}
