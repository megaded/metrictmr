package storage

import (
	"context"
	"reflect"
	"testing"

	"github.com/megaded/metrictmr/internal/data"
)

func TestInMemoryStorage_GetGauge(t *testing.T) {
	type args struct {
		name string
	}
	storeWithData := NewInMemoryStorage()
	metricName := "test"
	gauge := data.Metric{MType: data.MTypeGauge, ID: metricName}
	storeWithData.Store(context.TODO(), gauge)
	tests := []struct {
		name       string
		s          *InMemoryStorage
		args       args
		wantMetric data.Metric
		wantExist  bool
		wantErr    bool
	}{
		{"metric exists ", storeWithData, args{name: metricName}, gauge, true, false},
		{"metric not exists", NewInMemoryStorage(), args{name: "1"}, gauge, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetric, gotExist, _ := tt.s.GetGauge(tt.args.name)
			if !gotExist {
				return
			}
			if !reflect.DeepEqual(gotMetric, tt.wantMetric) {
				t.Errorf("InMemoryStorage.GetGauge() gotMetric = %v, want %v", gotMetric, tt.wantMetric)
			}
			if gotExist != tt.wantExist {
				t.Errorf("InMemoryStorage.GetGauge() gotExist = %v, want %v", gotExist, tt.wantExist)
			}
		})
	}
}

func TestInMemoryStorage_Store(t *testing.T) {
	type args struct {
		ctx    context.Context
		metric []data.Metric
	}
	tests := []struct {
		name    string
		s       *InMemoryStorage
		args    args
		wantErr bool
	}{
		{"store", NewInMemoryStorage(), args{ctx: context.TODO(), metric: make([]data.Metric, 1, 1)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Store(tt.args.ctx, tt.args.metric...); (err != nil) != tt.wantErr {
				t.Errorf("InMemoryStorage.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMemoryStorage_GetCounter(t *testing.T) {
	type args struct {
		name string
	}
	storeWithData := NewInMemoryStorage()
	metricName := "test"
	counter := data.Metric{MType: data.MTypeGauge, ID: metricName}
	storeWithData.Store(context.TODO(), counter)
	tests := []struct {
		name       string
		s          *InMemoryStorage
		args       args
		wantMetric data.Metric
		wantExist  bool
		wantErr    bool
	}{
		{"metric exists ", storeWithData, args{name: metricName}, counter, true, false},
		{"metric not exists", NewInMemoryStorage(), args{name: "1"}, counter, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetric, gotExist, _ := tt.s.GetCounter(tt.args.name)
			if gotExist {
				return
			}
			if !reflect.DeepEqual(gotMetric, tt.wantMetric) {
				t.Errorf("InMemoryStorage.GetCounter() gotMetric = %v, want %v", gotMetric, tt.wantMetric)
			}
			if gotExist != tt.wantExist {
				t.Errorf("InMemoryStorage.GetCounter() gotExist = %v, want %v", gotExist, tt.wantExist)
			}
		})
	}
}

func TestInMemoryStorage_GetMetrics(t *testing.T) {
	storeWithData := NewInMemoryStorage()
	metricName := "test"
	counter := data.Metric{MType: data.MTypeGauge, ID: metricName}
	storeWithData.Store(context.TODO(), counter)
	tests := []struct {
		name    string
		s       *InMemoryStorage
		want    []data.Metric
		wantErr bool
	}{
		{"get metrics ok", storeWithData, []data.Metric{counter}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetMetrics()
			if (err != nil) != tt.wantErr {
				t.Errorf("InMemoryStorage.GetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InMemoryStorage.GetMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
