package httpserver

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/ray-project/kuberay/ray-operator/controllers/ray/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"strings"
)

// validateAccess validates the access permissions for the auto scaler.
// This function checks the AccessKey and AccessSecret fields in the HTTP request headers.
// Parameters:
//
//	request (*http.Request): The HTTP request to be validated.
//
// Returns:
//
//	bool: Returns true if the access validation is successful, otherwise false.
//
// Caution:when used in production, you should replace this function with your actual access validation logic.
func validateAccess(request *http.Request, logger logr.Logger) bool {
	// Check if the AccessKey matches.
	if !strings.EqualFold(request.Header.Get("AccessKey"), "KubeRayOperator") {
		logger.Error(fmt.Errorf("parameter error"), "access key is not match")
		return false
	}

	// Check if the AccessSecret matches.
	if !strings.EqualFold(request.Header.Get("AccessSecret"), "S3ViZVJheU9wZXJhdG9y") {
		logger.Error(fmt.Errorf("parameter error"), "access secret is not match")
		return false
	}

	// If both AccessKey and AccessSecret match, return true.
	return true
}

func (s OperatorHttpServer) RayClusterGetHandler(writer http.ResponseWriter, request *http.Request) {

	if !validateAccess(request, s.logger) {
		json.NewEncoder(writer).Encode(utils.NewJsonResponse(false, fmt.Sprintf("AccessKey or AccessSecret is not match."), nil))
		return
	}

	// 从路由变量中获取动态部分
	vars := mux.Vars(request)
	k8sNamespace := vars["namespace"]
	clusterName := vars["cluster"]

	// 返回结果
	s.logger.Info("RayClusterGetHandler", "namespace", k8sNamespace, "cluster", clusterName)

	rayCluster, _ := s.KubeRayClient.RayV1().RayClusters(k8sNamespace).Get(request.Context(), clusterName, metav1.GetOptions{})
	json.NewEncoder(writer).Encode(rayCluster)

}

func (s OperatorHttpServer) RayClusterPodGetHandler(writer http.ResponseWriter, request *http.Request) {

	if !validateAccess(request, s.logger) {
		json.NewEncoder(writer).Encode(utils.NewJsonResponse(false, fmt.Sprintf("AccessKey or AccessSecret is not match."), nil))
		return
	}

	// 从路由变量中获取动态部分
	vars := mux.Vars(request)
	k8sNamespace := vars["namespace"]
	podname := vars["podname"]

	// 返回结果
	s.logger.Info("RayClusterPodGetHandler", "namespace", k8sNamespace, "podname", podname)

	pod, _ := s.Kubeclient.CoreV1().Pods(k8sNamespace).Get(request.Context(), podname, metav1.GetOptions{})
	json.NewEncoder(writer).Encode(pod)

}

func (s OperatorHttpServer) RayClusterPodListHandler(writer http.ResponseWriter, request *http.Request) {

	if !validateAccess(request, s.logger) {
		json.NewEncoder(writer).Encode(utils.NewJsonResponse(false, fmt.Sprintf("AccessKey or AccessSecret is not match."), nil))
		return
	}

	// 从路由变量中获取动态部分
	vars := mux.Vars(request)
	k8sNamespace := vars["namespace"]

	// 解析查询参数 labelSelector
	queryParams := request.URL.Query()
	labelSelector := queryParams.Get("labelSelector")

	// 返回结果
	s.logger.Info("RayClusterPodListHandler", "namespace", k8sNamespace, "cluster", labelSelector)

	podList, _ := s.Kubeclient.CoreV1().Pods(k8sNamespace).List(request.Context(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	json.NewEncoder(writer).Encode(podList)

}

func (s OperatorHttpServer) RayClusterPatchHandler(writer http.ResponseWriter, request *http.Request) {

	if !validateAccess(request, s.logger) {
		json.NewEncoder(writer).Encode(utils.NewJsonResponse(false, fmt.Sprintf("AccessKey or AccessSecret is not match."), nil))
		return
	}

	// 从路由变量中获取动态部分
	vars := mux.Vars(request)
	k8sNamespace := vars["namespace"]
	clusterName := vars["cluster"]

	patchData := make([]utils.JsonPatchData, 0)
	if err := json.NewDecoder(request.Body).Decode(&patchData); err != nil {
		respondWithError(writer, s.logger, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format:%s", err.Error()))
		return
	}

	// 返回结果
	s.logger.Info("RayClusterPatchHandler", "namespace", k8sNamespace, "cluster", clusterName, "patchData", patchData)
	playLoadBytes, _ := json.Marshal(patchData)
	rayCluster, _ := s.KubeRayClient.RayV1().RayClusters(k8sNamespace).Patch(request.Context(), clusterName, types.JSONPatchType, playLoadBytes, metav1.PatchOptions{})

	json.NewEncoder(writer).Encode(rayCluster)

}

// 辅助函数：返回错误响应
func respondWithError(w http.ResponseWriter, logger logr.Logger, code int, message string) {
	respondWithJSON(w, logger, code, map[string]string{"error": message})
}

// 辅助函数：返回 JSON 响应
func respondWithJSON(w http.ResponseWriter, logger logr.Logger, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Info("Error encoding JSON response:", "error", err)
	}
}
