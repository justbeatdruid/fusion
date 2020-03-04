package user

const (
	approverLabel  = "nlpt.cmcc.com/approver"
	applicantLabel = "nlpt.cmcc.com/applicant"
)

type ApplyUser struct {
	ApprovedBy User `json:"approver"`
	AppliedBy  User `json:"applicant"`
}

func AddApplyLabel(au ApplyUser, labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[approverLabel] = au.ApprovedBy.ID
	labels[applicantLabel] = au.AppliedBy.ID
	return labels
}

func InitWithApplicant(id string) ApplyUser {
	return ApplyUser{
		AppliedBy: User{
			ID: id,
		},
	}
}

func GetApplyUserFromLabels(labels map[string]string) ApplyUser {
	au := ApplyUser{
		ApprovedBy: User{
			ID: labels[approverLabel],
		},
		AppliedBy: User{
			ID: labels[applicantLabel],
		},
	}
	au.ApprovedBy.Name = idnames(au.ApprovedBy.ID)
	au.AppliedBy.Name = idnames(au.AppliedBy.ID)
	return au
}

func GetApproverLabelSelector(id string) string {
	return approverLabel + "=" + id
}

func GetApplicantLabelSelector(id string) string {
	return applicantLabel + "=" + id
}
