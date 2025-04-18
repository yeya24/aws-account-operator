package awsfederatedrole

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awsv1alpha1 "github.com/openshift/aws-account-operator/api/v1alpha1"
	"github.com/openshift/aws-account-operator/pkg/utils"
)

func (r *AWSFederatedRoleReconciler) addFinalizer(reqLogger logr.Logger, awsFederatedRole *awsv1alpha1.AWSFederatedRole) error {
	reqLogger.Info("Adding Finalizer for the AccountClaim")
	awsFederatedRole.SetFinalizers(append(awsFederatedRole.GetFinalizers(), utils.Finalizer))

	// Update CR
	err := r.Update(context.TODO(), awsFederatedRole)
	if err != nil {
		reqLogger.Error(err, "Failed to update AccountClaim with finalizer")
		return err
	}
	return nil
}

func (r *AWSFederatedRoleReconciler) removeFinalizer(reqLogger logr.Logger, awsFederatedRole *awsv1alpha1.AWSFederatedRole, finalizerName string) error {
	reqLogger.Info("Removing Finalizer for the AWSFederatedRole")
	awsFederatedRole.SetFinalizers(utils.Remove(awsFederatedRole.GetFinalizers(), finalizerName))

	// Update CR
	err := r.Update(context.TODO(), awsFederatedRole)
	if err != nil {
		reqLogger.Error(err, "Failed to remove AWSFederatedAccountAccess finalizer")
		return err
	}
	return nil
}

func (r *AWSFederatedRoleReconciler) finalizeFederateRole(reqLogger logr.Logger, awsFederatedRole *awsv1alpha1.AWSFederatedRole) error {

	// Get all FederatedAccountAccesses
	awsFederatedAccountAccessList := &awsv1alpha1.AWSFederatedAccountAccessList{}

	listOpts := []client.ListOption{}
	if err := r.List(context.TODO(), awsFederatedAccountAccessList, listOpts...); err != nil {
		reqLogger.Error(err, "unable to list AWS Federated Account Accesses")
		return err
	}

	for i := range awsFederatedAccountAccessList.Items {
		if isFederatedRoleReferenced(&awsFederatedAccountAccessList.Items[i], awsFederatedRole) {
			deleteAccessErr := r.Delete(context.TODO(), &awsFederatedAccountAccessList.Items[i])
			if deleteAccessErr != nil {
				reqLogger.Error(deleteAccessErr, fmt.Sprintf("unable to delete AWS Federated Account Accesses %s\n", awsFederatedAccountAccessList.Items[i].Name))
				return deleteAccessErr
			}
		}

	}

	return nil
}

func isFederatedRoleReferenced(awsFederatedAccountAccess *awsv1alpha1.AWSFederatedAccountAccess, awsFederatedRole *awsv1alpha1.AWSFederatedRole) bool {

	referencedRoleNamespacedName := types.NamespacedName{Name: awsFederatedAccountAccess.Spec.AWSFederatedRole.Name, Namespace: awsFederatedAccountAccess.Spec.AWSFederatedRole.Namespace}
	roleNamespacedName := types.NamespacedName{Name: awsFederatedRole.Name, Namespace: awsFederatedRole.Namespace}

	return reflect.DeepEqual(referencedRoleNamespacedName, roleNamespacedName)
}
