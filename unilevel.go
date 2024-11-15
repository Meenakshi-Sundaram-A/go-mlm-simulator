package main

type UniMember struct {
	ID           int
	Parent       *UniMember
	Children     []*UniMember
	Level        int
	Sale         float64
	SponsorBonus float64
	// BinaryBonus          float64
	// DownlineSales        float64
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

func NewUniLevelTree(numMembers int, packagePrice float64, maxChild int) *UniLevelTree {
	tree := &UniLevelTree{
		NumMembers:   numMembers,
		PackagePrice: packagePrice,
	}
	tree.buildUniLevelTree(maxChild)
	tree.setUniLevelMemberSales(packagePrice)
	return tree
}

func (t *UniLevelTree) buildUniLevelTree(maxChild int) {
	if t.NumMembers <= 0 {
		return
	}
	t.Root = &UniMember{ID: 1, Level: 1}
	t.Members = append(t.Members, t.Root)
	queue := []*UniMember{t.Root}
	currentID := 2

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
func (t *UniLevelTree) setUniLevelMemberSales(packagePrice float64) {
	for _, member := range t.Members {
		if member.ID != 1 {
			member.Sale = packagePrice
		}
	}
}

func (t *UniLevelTree) unilevelSponsorBonus(sponsorPercentage float64, packagePrice float64, cappingAmount float64) float64 {
	var totalSponsorBonus float64

	for _, member := range t.Members {
		if len(member.Children) != 0 {
			childCount := float64(len(member.Children))
			member.SponsorBonus = childCount * packagePrice * (sponsorPercentage / 100)

			if cappingAmount > 0 && member.SponsorBonus > cappingAmount {
				member.SponsorBonus = cappingAmount
			}

			totalSponsorBonus += member.SponsorBonus
		}
	}
	return totalSponsorBonus
}

// func sendResultsToDjango(results interface{}) {
// 	jsonData, err := json.Marshal(results)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	resp, err := http.Post("http://localhost:8000/process_results/", "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()
// 	fmt.Println("Response from Django:", resp.Status)
// }

func convertToUniLevelJSONStructure(members []*UniMember) []map[string]interface{} {
	var jsonNodes []map[string]interface{}
	for _, member := range members {
		parentID := 0
		if member.Parent != nil {
			parentID = member.Parent.ID
		}

		jsonNodes = append(jsonNodes, map[string]interface{}{
			"ID":           member.ID,
			"Level":        member.Level,
			"ParentID":     parentID,
			"SponsorBonus": member.SponsorBonus,
		})
	}
	return jsonNodes
}

func ProcessUnilevelTree(data map[string]interface{}) map[string]interface{} {
	numOfUsers := data["num_of_users"].(float64)
	packagePrice := data["package_price"].(float64)
	sponsorBonusPercentage := data["sponsor_bonus_percentage"].(float64)
	maxChild := data["max_child"].(float64)
	cappingAmount := data["capping_amount"].(float64)

	// cappingScope := data["capping_scope"].(string)
	// lev1Percentage := data["lev1_percentage"].(float64)
	// lev2Percentage := data["lev2_percentage"].(float64)

	tree := NewUniLevelTree(int(numOfUsers), packagePrice, int(maxChild))
	sponsorBonus := tree.unilevelSponsorBonus(sponsorBonusPercentage, packagePrice, cappingAmount)
	// totalBinaryBonus := tree.setBinaryBonus(binaryBonusPercentage, cappingAmount)
	// totalMatchingBonus := tree.setMatchingBonus(lev1Percentage, lev2Percentage)

	return map[string]interface{}{
		"tree_structure":      convertToUniLevelJSONStructure(tree.Members),
		"total_sponsor_bonus": sponsorBonus,
		// "total_binary_bonus":   totalBinaryBonus,
		// "total_matching_bonus": totalMatchingBonus,
	}
}
