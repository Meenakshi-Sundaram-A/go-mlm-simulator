package main

type UniMember struct {
	ID            int
	Parent        *UniMember
	Children      []*UniMember
	PackagePrice  float64
	Level         int
	Sale          float64
	SponsorBonus  float64
	MatchingBonus float64
	// DownlineSales        float64
	// CarryForward         float64
	// CarryForwardPosition string
}

type UniLevelTree struct {
	Root          *UniMember
	NumMembers    int
	ProductsPrice []float64
	Members       []*UniMember
}

func NewUniLevelTree(numMembers int, productsPrice []float64, maxChild int) *UniLevelTree {
	tree := &UniLevelTree{
		NumMembers:    numMembers,
		ProductsPrice: productsPrice,
	}
	// tree.buildUniLevelTree(maxChild)
	// tree.setUniLevelMemberSales(packagePrice)
	return tree
}

func sumUsers(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

func (t *UniLevelTree) buildUniLevelTree(maxChild int, usersPerProduct []float64, queue []*UniMember) []*UniMember {
	currCount := 0
	if t.NumMembers <= 0 {
		return queue
	}

	totalUsersPerCycle := sumUsers(usersPerProduct)
	currentID := queue[len(queue)-1].ID + 1

	for currentID <= t.NumMembers && currCount < int(totalUsersPerCycle) {

		if len(queue) == 0 {
			break
		}

		currentMember := queue[0]
		flag := false

		if currentID <= t.NumMembers && len(currentMember.Children) != maxChild {
			for index := range usersPerProduct {
				if usersPerProduct[index] > 0 {
					newChild := &UniMember{ID: currentID, Parent: currentMember, Level: currentMember.Level + 1, PackagePrice: t.ProductsPrice[index]}
					currentMember.Children = append(currentMember.Children, newChild)
					queue = append(queue, newChild)
					t.Members = append(t.Members, newChild)
					usersPerProduct[index]--
					currCount++
					currentID++
					break
				}
			}
		}
		if len(currentMember.Children) == maxChild {
			flag = true
		}
		if flag {
			queue = queue[1:]
		}
	}
	return queue
}

// func (t *UniLevelTree) buildUniLevelTree(maxChild int, usersPerProduct []float64, queue []*UniMember) []*UniMember {
// 	currCount := 0
// 	if t.NumMembers <= 0 {
// 		return queue
// 	}

// 	totalUsersPerCycle := sumUsers(usersPerProduct)
// 	currentID :=  queue[len(queue)-1].ID + 1

//		for currentID <= t.NumMembers && currCount < int(totalUsersPerCycle) {
//			currentMember := queue[0]
//			flag := false
//			if len(currentMember.Children) != maxChild {
//				newChild := &UniMember{ID: currentID, Parent: currentMember, Level: currentMember.Level + 1}
//				currentMember.Children = append(currentMember.Children, newChild)
//				queue = append(queue, newChild)
//				t.Members = append(t.Members, newChild)
//				currentID++
//			} else {
//				flag = true
//			}
//			if flag {
//				queue = queue[1:]
//			}
//		}
//		return queue
//	}

func (t *UniLevelTree) unilevelSponsorBonus(sponsorPercentage float64, cappingAmount float64, cappingScope []string) (float64, float64) {
	var totalSponsorBonus float64
	var revenue float64

	flag := false
	for _, item := range cappingScope {
		if item == "sponsor_bonus" {
			flag = true
		}
	}

	for _, member := range t.Members {
		sponsorBonus := 0.0
		if len(member.Children) != 0 {
			for _, child := range member.Children {
				sponsorBonus += child.PackagePrice * (sponsorPercentage / 100)
				revenue += child.PackagePrice

			}
			if flag && cappingAmount > 0 && sponsorBonus > cappingAmount {
				member.SponsorBonus = cappingAmount
			} else {
				member.SponsorBonus = sponsorBonus
			}
		}
		totalSponsorBonus += member.SponsorBonus
	}
	return totalSponsorBonus, revenue
}

func (t *UniLevelTree) unilevelMatchingBonus(levelPercentages []float64, cappingAmount float64, cappingScope []string) float64 {
	totalMatchingBonus := 0.0

	flag := false
	for _, item := range cappingScope {
		if item == "sponsor_bonus" {
			flag = true
		}
	}

	for _, member := range t.Members {
		queue := []*UniMember{member}
		member.MatchingBonus = 0.0
		for _, level := range levelPercentages {
			nextLevelNodes := []*UniMember{}

			for _, node := range queue {
				for _, child := range node.Children {
					nextLevelNodes = append(nextLevelNodes, child)
					member.MatchingBonus += child.SponsorBonus * (level / 100)
					if flag && cappingAmount > 0 && member.MatchingBonus > cappingAmount {
						member.MatchingBonus = cappingAmount
					}
				}
			}
			queue = nextLevelNodes
		}
		totalMatchingBonus += member.MatchingBonus
	}
	return totalMatchingBonus
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

func convertToUniLevelJSONStructureForAdmin(members []*UniMember) []map[string]interface{} {
	var jsonNodes []map[string]interface{}
	for _, member := range members {
		if member.ID == 1 {
			jsonNodes = append(jsonNodes, map[string]interface{}{
				"ID":            member.ID,
				"SponsorBonus":  member.SponsorBonus,
				"MatchingBonus": member.MatchingBonus,
			})
			break
		}
	}
	return jsonNodes
}

func convertToUniLevelJSONStructure(members []*UniMember) []map[string]interface{} {
	var jsonNodes []map[string]interface{}
	for _, member := range members {
		parentID := 0
		if member.Parent != nil {
			parentID = member.Parent.ID
		}

		jsonNodes = append(jsonNodes, map[string]interface{}{
			"ID":            member.ID,
			"Level":         member.Level,
			"ParentID":      parentID,
			"PackagePrice":  member.PackagePrice,
			"SponsorBonus":  member.SponsorBonus,
			"MatchingBonus": member.MatchingBonus,
		})
	}
	return jsonNodes
}

func ProcessUnilevelTree(data map[string]interface{}) []map[string]interface{} {

	numOfUsers := int(data["num_of_users"].(float64)) + 1
	cycles := int(data["cycle"].(float64))
	sponsorBonusPercentage := data["sponsor_bonus_percentage"].(float64)
	poolBonusPercentage := data["pool_bonus_percentage"].(float64)
	poolBonusCount := data["pool_bonus_count"].(float64)

	matchingBonusPercentages := []float64{}
	if rawPercentages, ok := data["percentage_string"].([]interface{}); ok {
		for _, val := range rawPercentages {
			matchingBonusPercentages = append(matchingBonusPercentages, val.(float64))
		}
	}

	productsPrice := []float64{}
	if rawPercentages, ok := data["product_price"].([]interface{}); ok {
		for _, val := range rawPercentages {
			productsPrice = append(productsPrice, val.(float64))
		}
	}

	usersPerProduct := []float64{}
	if rawPercentages, ok := data["users_per_product"].([]interface{}); ok {
		for _, val := range rawPercentages {
			usersPerProduct = append(usersPerProduct, val.(float64))
		}
	}

	maxChild := int(data["max_child"].(float64))

	rawCappingScope := data["capping_scope"].([]interface{})
	cappingScope := make([]string, len(rawCappingScope))
	for i, v := range rawCappingScope {
		// Assert each element as a string
		if str, ok := v.(string); ok {
			cappingScope[i] = str
		}
	}

	cappingAmount := data["capping_amount"].(float64)

	tree := NewUniLevelTree(numOfUsers, productsPrice, int(maxChild))
	tree.Root = &UniMember{ID: 1, Level: 1}
	tree.Members = append(tree.Members, tree.Root)
	queue := []*UniMember{tree.Root}

	var totalSponsorBonus = 0.0
	var totalMatchingBonus = 0.0
	var revenue = 0.0
	var expense = 0.0
	var profit = 0.0
	var poolBonus = 0.0
	var results []map[string]interface{}
	for i := 0; i < cycles; i++ {
		usersPerProduct := []float64{}
		if rawPercentages, ok := data["users_per_product"].([]interface{}); ok {
			for _, val := range rawPercentages {
				usersPerProduct = append(usersPerProduct, val.(float64))
			}
		}
		queue = tree.buildUniLevelTree(int(maxChild), usersPerProduct, queue)
		totalSponsorBonus, revenue = tree.unilevelSponsorBonus(sponsorBonusPercentage, cappingAmount, cappingScope)
		totalMatchingBonus = tree.unilevelMatchingBonus(matchingBonusPercentages, cappingAmount, cappingScope)

		adminList := convertToUniLevelJSONStructureForAdmin(tree.Members)
		adminMatchingBonus := adminList[0]["MatchingBonus"].(float64)
		adminSponsorBonus := adminList[0]["SponsorBonus"].(float64)

		totalMatchingBonus = totalMatchingBonus - adminMatchingBonus
		totalSponsorBonus = totalSponsorBonus - adminSponsorBonus

		expense = totalSponsorBonus + totalMatchingBonus
		// fmt.Println("Expense:", expense)
		profit = revenue - expense
		if poolBonusPercentage > 0 && poolBonusCount > 0 {
			poolBonus = profit * (poolBonusPercentage / 100)
			profit = profit - poolBonus
			expense += poolBonus
		}

		ans := map[string]interface{}{
			// "tree_structure":       convertToUniLevelJSONStructure(tree.Members),
			"revenue":              revenue,
			"expense":              expense,
			"profit":               profit,
			"pool_bonus":           poolBonus,
			"total_sponsor_bonus":  totalSponsorBonus,
			"total_matching_bonus": totalMatchingBonus,
		}
		results = append(results, ans)
	}
	// for i, value := range results {
	// 	fmt.Print(i, " ", value)
	// }
	return results

	// tree := NewUniLevelTree(int(numOfUsers), packagePrice, int(maxChild))
	// sponsorBonus := tree.unilevelSponsorBonus(sponsorBonusPercentage, packagePrice, cappingAmount)
	// totalMatchingBonus := tree.unilevelMatchingBonus(matchingBonusPercentages)

	// return map[string]interface{}{
	// 	"tree_structure":       convertToUniLevelJSONStructure(tree.Members),
	// 	"total_sponsor_bonus":  sponsorBonus,
	// 	"total_matching_bonus": totalMatchingBonus,
	// 	// "total_binary_bonus":   totalBinaryBonus,
	// }
}
