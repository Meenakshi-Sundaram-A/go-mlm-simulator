package main

type UnilevelFormData struct {
	NumOfUsers             int     `json:"num_of_users"`
	PackagePrice           float64 `json:"package_price"`
	SponsorBonusPercentage int     `json:"sponsor_bonus_percentage"`
	BinaryBonusPercentage  int     `json:"binary_bonus_percentage"`
	Lev1Percentage         float64 `json:"lev1_percentage"`
	Lev2Percentage         float64 `json:"lev2_percentage"`
	CappingScope           string  `json:"capping_scope"`
	CappingAmount          int     `json:"capping_amount"`
	CarryYesNo             string  `json:"carry_yes_no"`
}

type UniMember struct {
	ID       int
	Parent   *UniMember
	Children []*UniMember
	Level    int
	// Sale                 float64
	// SponsorBonus         float64
	// BinaryBonus          float64
	// LeftSales            float64
	// RightSales           float64
	// CarryForward         float64
	// CarryForwardPosition string
	// MatchingBonus        float64
}

type UniLevelTree struct {
	Root         *UniMember
	NumMembers   int
	PackagePrice float64
	Members      []*UniMember
}

func NewUniLevelTree(numMembers int) *UniLevelTree {
	tree := &UniLevelTree{
		NumMembers: numMembers,
	}
	tree.buildUniLevelTree()
	// tree.setMemberSales(packagePrice)
	return tree
}

func (t *UniLevelTree) buildUniLevelTree() {
	if t.NumMembers <= 0 {
		return
	}
	t.Root = &UniMember{ID: 1, Level: 1}
	t.Members = append(t.Members, t.Root)
	queue := []*UniMember{t.Root}
	currentID := 2
	maxChild := 4

	for currentID <= t.NumMembers {
		currentMember := queue[0]
		if len(currentMember.Children) != maxChild {
			newChild := &UniMember{ID: currentID, Parent: currentMember, Level: currentMember.Level + 1}
			currentMember.Children = append(currentMember.Children, newChild)
			queue = append(queue, newChild)
			t.Members = append(t.Members, newChild)
			currentID++
		} else {
			queue = queue[1:]
		}
	}
}

func convertToUniLevelJSONStructure(members []*UniMember) []map[string]interface{} {
	var jsonNodes []map[string]interface{}
	for _, member := range members {
		parentID := 0
		if member.Parent != nil {
			parentID = member.Parent.ID
		}

		jsonNodes = append(jsonNodes, map[string]interface{}{
			"ID":       member.ID,
			"Level":    member.Level,
			"ParentID": parentID,
		})
	}
	return jsonNodes
}

func ProcessUnilevelTree(data map[string]interface{}) map[string]interface{} {
	numOfUsers := int(data["num_of_users"].(float64))
	// packagePrice := data["package_price"].(float64)
	// sponsorBonusPercentage := data["sponsor_bonus_percentage"].(float64)
	// binaryBonusPercentage := data["binary_bonus_percentage"].(float64)
	// lev1Percentage := data["lev1_percentage"].(float64)
	// lev2Percentage := data["lev2_percentage"].(float64)
	// cappingScope := data["capping_scope"].(string)
	// cappingAmount := data["capping_amount"].(float64)

	tree := NewUniLevelTree(numOfUsers)
	// sponsorBonus := tree.setAndGetSponsorBonus(sponsorBonusPercentage, cappingAmount, cappingScope)
	// totalBinaryBonus := tree.setBinaryBonus(binaryBonusPercentage, cappingAmount)
	// totalMatchingBonus := tree.setMatchingBonus(lev1Percentage, lev2Percentage)

	return map[string]interface{}{
		"tree_structure": convertToUniLevelJSONStructure(tree.Members),
		// "total_sponsor_bonus":  sponsorBonus,
		// "total_binary_bonus":   totalBinaryBonus,
		// "total_matching_bonus": totalMatchingBonus,
	}
}
