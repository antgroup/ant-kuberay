package httpserver

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"net/http"
	_ "net/http/pprof"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	"github.com/ray-project/kuberay/ray-operator/pkg/client/clientset/versioned"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func StartServer(ctx context.Context, config *rest.Config, kubeclient kubernetes.Interface) error {

	operatorHttpServer := OperatorHttpServer{
		Kubeclient:    kubeclient,
		KubeRayClient: versioned.NewForConfigOrDie(config),
		logger:        ctrl.LoggerFrom(ctx).WithName("kuberay-http-server"),
	}

	// 创建路由实例
	router := mux.NewRouter()

	// antray autoscaler provider
	router.HandleFunc("/apis/ray.io/v1/namespaces/{namespace}/rayclusters/{cluster}", operatorHttpServer.RayClusterGetHandler).Methods("GET")
	// /api/v1/namespaces/{namespace}/pods?labelSelector=ray.io/cluster={cluster}
	router.HandleFunc("/api/v1/namespaces/{namespace}/pods", operatorHttpServer.RayClusterPodListHandler).Methods("GET")
	router.HandleFunc("/api/v1/namespaces/{namespace}/pods/{podname}", operatorHttpServer.RayClusterPodGetHandler).Methods("GET")
	router.HandleFunc("/apis/ray.io/v1/namespaces/{namespace}/rayclusters/{cluster}", operatorHttpServer.RayClusterPatchHandler).Methods("PATCH")

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 40 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      router,
	}
	operatorHttpServer.logger.Info("web server start" + srv.Addr)
	err := srv.ListenAndServe()
	if err != nil {
		operatorHttpServer.logger.Error(err, "http server down")
		return err
	}
	return nil
}

type OperatorHttpServer struct {
	Kubeclient    kubernetes.Interface
	KubeRayClient versioned.Interface
	logger        logr.Logger
}
