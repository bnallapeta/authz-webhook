package webhook

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type UserValidator struct {
	Client  kubernetes.Interface
	decoder *admission.Decoder
	logger  *zap.Logger
}

// NewUserValidator constructs a new UserValidator
func NewUserValidator() (*UserValidator, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/bnr/.kube/kind.kubeconfig")
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	logger = logger.WithOptions(zap.AddCallerSkip(1))

	logger.Info("Testing if this is getting printed 2")

	return &UserValidator{
		Client: client,
		logger: logger,
	}, nil
}

// InjectDecoder injects the decoder into UserValidator
func (v *UserValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// Handle handles HTTP requests
func (v *UserValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	v.logger.Info("Inside Handle function",
		zap.String("User", req.UserInfo.Username),
		zap.String("APIVersion", req.Kind.Version),
		zap.String("Kind", req.Kind.Kind),
	)

	v.logger.Info("Testing if we can fetch pods")

	pods, err := v.Client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, pod := range pods.Items {
		v.logger.Sugar().Infof("Pod Name: %s\n", pod.Name)
	}

	// sar := &authv1.SubjectAccessReview{
	// 	Spec: authv1.SubjectAccessReviewSpec{
	// 		User: req.UserInfo.Username,
	// 		ResourceAttributes: &authv1.ResourceAttributes{
	// 			Namespace: req.Namespace,
	// 			Verb:      string(req.Operation),
	// 			Group:     req.Kind.Group,
	// 			Version:   req.Kind.Version,
	// 			Resource:  "pods",
	// 		},
	// 	},
	// }

	sar := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			User: "kubernetes-admin",
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: "default",
				Verb:      "create",
				Group:     "core",
				Version:   "v1",
				Resource:  "pods",
			},
		},
	}

	v.logger.Info("sar object", zap.Any("sar", sar))

	// Create the SubjectAccessReview object
	res, err := v.Client.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		v.logger.Error("Failed to create SubjectAccessReview", zap.Error(err))
		return admission.Errored(http.StatusInternalServerError, err)
	}

	v.logger.Info("SAR after create call", zap.Any("res", res))

	if res.Status.Allowed {
		v.logger.Info("Access allowed")
		return admission.Allowed("Access allowed")
	}
	v.logger.Info("Access denied", zap.String("Reason", res.Status.Reason))
	return admission.Denied(res.Status.Reason)
}
