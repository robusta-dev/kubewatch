package utils

import (
	"testing"
	"time"

	apps_v1 "k8s.io/api/apps/v1"
	autoscaling_v1 "k8s.io/api/autoscaling/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestGetObjectMetaData covers resource types the controller watches. The
// create path compares CreationTimestamp with the controller's start time, so a
// type missing here never produces a create notification.
func TestGetObjectMetaData(t *testing.T) {
	want := meta_v1.ObjectMeta{
		Name:              "web",
		Namespace:         "default",
		CreationTimestamp: meta_v1.NewTime(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)),
	}

	tests := map[string]interface{}{
		"StatefulSet":             &apps_v1.StatefulSet{ObjectMeta: want},
		"HorizontalPodAutoscaler": &autoscaling_v1.HorizontalPodAutoscaler{ObjectMeta: want},
	}

	for name, obj := range tests {
		t.Run(name, func(t *testing.T) {
			got := GetObjectMetaData(obj)
			if got.Name != want.Name || got.Namespace != want.Namespace || !got.CreationTimestamp.Equal(&want.CreationTimestamp) {
				t.Errorf("GetObjectMetaData() = %s/%s created %v, want %s/%s created %v",
					got.Namespace, got.Name, got.CreationTimestamp, want.Namespace, want.Name, want.CreationTimestamp)
			}
		})
	}
}
