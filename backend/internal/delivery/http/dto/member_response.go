package dto

type AddMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

type AddMembersResponse struct {
	AddedUserIDs []string `json:"added_user_ids"`
	AddedCount   int      `json:"added_count"`
}

func NewAddMembersResponse(userIDs []string) AddMembersResponse {
	if userIDs == nil {
		userIDs = []string{}
	}
	return AddMembersResponse{AddedUserIDs: userIDs, AddedCount: len(userIDs)}
}
