package expect

import (
	"github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Error expects an error happens, otherwise an exception raises
func Error(err error, explain ...any) {
	gomega.ExpectWithOffset(1, err).To(gomega.HaveOccurred(), explain...)
}

// NoError checks if "err" is set and raises an exception if so
func NoError(err error, explain ...any) {
	gomega.ExpectWithOffset(1, err, explain...).ShouldNot(gomega.HaveOccurred())
}

// Forbidden checks if "err" is set a `metav1.StatusReasonForbidden` or `404`
// response using `apierrors` package, raising an exception otherwise
func Forbidden(err error) {
	gomega.ExpectWithOffset(1, apierrors.IsForbidden(err)).To(gomega.BeTrue(), "expected to be forbidden: %s", err)
}
