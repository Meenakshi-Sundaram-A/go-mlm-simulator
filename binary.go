// binary.go
package main

import (
	"fmt"
	"math"
)

type Member struct {
	ID            int
	Parent        *Member
	LeftMember    *Member
	RightMember   *Member
	Position      string
	Level         int
	PackagePrice  float64
	Sale          float64
	SponsorBonus  float64
	BinaryBonus   float64
	MatchingBonus float64
	LeftSales     float64
	RightSales    float64
	LeftCarry     float64
	RightCarry    float64
}

type Tree struct {
	Root          *Member
	NumMembers    int
	ProductsPrice []float64
	Members       []*Member
}

func NewTree(numMembers int, productsPrice []float64) *Tree {
	tree := &Tree{
		NumMembers:    numMembers,
		ProductsPrice: productsPrice,
	}
	return tree
}

func sumSlice(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

func (t *Tree) buildTree(usersPerProduct []float64, queue []*Member) []*Member {
	currCount := 0
	if t.NumMembers <= 0 {
		return queue
	}
	totalUsersPerCycle := sumSlice(usersPerProduct)
	currId := queue[len(queue)-1].ID + 1

	for currId <= t.NumMembers && currCount < int(totalUsersPerCycle) {
		if len(queue) == 0 {
			break
		}
		currMember := queue[0]

		flag := false
		if currId <= t.NumMembers && currMember.LeftMember == nil {
			for index := range usersPerProduct {
				if usersPerProduct[index] > 0 {
					leftChild := &Member{ID: currId, Parent: currMember, Position: "Left", Level: currMember.Level + 1, PackagePrice: t.ProductsPrice[index]}
					currMember.LeftMember = leftChild
					queue = append(queue, leftChild)
					t.Members = append(t.Members, leftChild)
					usersPerProduct[index]--
					currCount++
					currId++
					break
				}
			}
		}

		if currId <= t.NumMembers {
			for index := range usersPerProduct {
				if usersPerProduct[index] > 0 {
					rightChild := &Member{ID: currId, Parent: currMember, Position: "Right", Level: currMember.Level + 1, PackagePrice: t.ProductsPrice[index]}
					currMember.RightMember = rightChild
					queue = append(queue, rightChild)
					t.Members = append(t.Members, rightChild)
					usersPerProduct[index]--
					currCount++
					currId++
					flag = true
					break
				}
			}
		}
		if flag {
			queue = queue[1:]
		}
	}
	return queue
}

func (t *Tree) setAndGetSponsorBonus(sponsorPercentage, cappingAmount float64, cappingScope []string) float64 {
	totalBonus := 0.0
	flag := false
	for _, item := range cappingScope {
		if item == "sponsor_bonus" {
			flag = true
		}
	}
	for _, member := range t.Members {
		rightBonus := 0.0
		leftBonus := 0.0
		if member.RightMember != nil {
			rightBonus = member.RightMember.PackagePrice * (sponsorPercentage / 100)
		}
		if member.LeftMember != nil {
			leftBonus = member.LeftMember.PackagePrice * (sponsorPercentage / 100)
		}
		sponsorBonus := rightBonus + leftBonus
		if cappingAmount > 0 && flag && sponsorBonus > cappingAmount {
			member.SponsorBonus = cappingAmount
		} else {
			member.SponsorBonus = sponsorBonus
		}
		totalBonus += member.SponsorBonus
	}
	return totalBonus
}

func (t *Tree) setBinaryBonus(cappingAmount float64, leftRatioAmount float64, rightRatioAmount float64, cappingScope []string, binaryBonusPercentage float64) (float64, float64) {
	totalBonus := 0.0
	cache := make(map[*Member]float64)
	revenue := 0.0
	flag := false
	for _, item := range cappingScope {
		if item == "binary_bonus" {
			flag = true
		}
	}

	for _, member := range t.Members {
		leftSales := 0.0
		rightSales := 0.0

		if member.LeftMember != nil {
			leftSales = t.traverse(member.LeftMember, cache)
			if member.ID == 1 {
				revenue += leftSales
			}
			if member.LeftCarry > 0.0 {
				leftSales += member.LeftCarry
				member.LeftCarry = 0.0
			}
			member.LeftSales = leftSales
		}

		if member.RightMember != nil {
			rightSales = t.traverse(member.RightMember, cache)
			if member.ID == 1 {
				revenue += rightSales
			}
			if member.RightCarry > 0.0 {
				rightSales += member.RightCarry
				member.RightCarry = 0.0
			}
			member.RightSales = rightSales
		}

		miniVal := math.Min(member.LeftSales, member.RightSales)
		if binaryBonusPercentage > 0.0 {
			binaryBonus := miniVal * (binaryBonusPercentage / 100)
			if flag && cappingAmount > 0 && binaryBonus > cappingAmount {
				member.BinaryBonus = cappingAmount
			} else {
				member.BinaryBonus = binaryBonus
			}
			member.LeftCarry = member.LeftSales - miniVal
			member.RightCarry = member.RightSales - miniVal
		} else {
			pairCount := int(math.Min((leftSales / leftRatioAmount), (rightSales / rightRatioAmount)))
			leftVal := float64(pairCount) * leftRatioAmount
			rightVal := float64(pairCount) * rightRatioAmount
			minValue := math.Min(leftVal, rightVal)

			if pairCount <= 5 {
				binaryBonus := minValue * (10.0 / 100)
				if flag && cappingAmount > 0 && binaryBonus > cappingAmount {
					member.BinaryBonus = cappingAmount
				} else {
					member.BinaryBonus = binaryBonus
				}
			} else if pairCount > 5 && pairCount <= 10 {
				binaryBonus := minValue * (15.0 / 100)
				if flag && cappingAmount > 0 && binaryBonus > cappingAmount {
					member.BinaryBonus = cappingAmount
				} else {
					member.BinaryBonus = binaryBonus
				}
			} else if pairCount > 10 {
				binaryBonus := minValue * (20.0 / 100)
				if flag && cappingAmount > 0 && binaryBonus > cappingAmount {
					member.BinaryBonus = cappingAmount
				} else {
					member.BinaryBonus = binaryBonus
				}
			}
			member.LeftCarry = leftSales - (float64(pairCount) * leftRatioAmount)
			member.RightCarry = rightSales - (float64(pairCount) * rightRatioAmount)
		}

		totalBonus += member.BinaryBonus
	}
	return totalBonus, revenue
}

func (t *Tree) traverse(node *Member, cache map[*Member]float64) float64 {
	if node == nil {
		return 0
	}

	if val, exists := cache[node]; exists {
		return val
	}

	currentSales := node.PackagePrice
	leftSales := t.traverse(node.LeftMember, cache)
	rightSales := t.traverse(node.RightMember, cache)

	totalSales := currentSales + leftSales + rightSales

	cache[node] = totalSales

	return totalSales
}

func (t *Tree) setMatchingBonus(levelPercentages []float64, cappingScope []string, cappingAmount float64) float64 {
	totalMatchingBonus := 0.0
	flag := false
	for _, item := range cappingScope {
		if item == "matching_bonus" {
			flag = true
		}
	}

	for _, member := range t.Members {
		member.MatchingBonus = 0.0
		queue := []*Member{member}

		for _, percentage := range levelPercentages {
			nextLevelNodes := []*Member{}

			for _, node := range queue {
				if node.LeftMember != nil {
					member.MatchingBonus += node.LeftMember.BinaryBonus * (percentage / 100)
					if flag && cappingAmount > 0 && member.MatchingBonus > cappingAmount {
						member.MatchingBonus = cappingAmount
					}
					nextLevelNodes = append(nextLevelNodes, node.LeftMember)
				}
				if node.RightMember != nil {
					member.MatchingBonus += node.RightMember.BinaryBonus * (percentage / 100)
					if flag && cappingAmount > 0 && member.MatchingBonus > cappingAmount {
						member.MatchingBonus = cappingAmount
					}
					nextLevelNodes = append(nextLevelNodes, node.RightMember)
				}
			}
			queue = nextLevelNodes

			if len(queue) == 0 {
				break
			}
		}
		totalMatchingBonus += member.MatchingBonus
	}
	return totalMatchingBonus
}

func convertToJSONStructureForAdmin(members []*Member) []map[string]interface{} {
	var jsonNodes []map[string]interface{}
	for _, member := range members {
		if member.ID == 1 {
			jsonNodes = append(jsonNodes, map[string]interface{}{
				"ID":            member.ID,
				"SponsorBonus":  member.SponsorBonus,
				"BinaryBonus":   member.BinaryBonus,
				"MatchingBonus": member.MatchingBonus,
			})
			break
		}
	}
	return jsonNodes
}

// func convertToJSONStructure(members []*Member) []map[string]interface{} {
// 	var jsonNodes []map[string]interface{}
// 	for _, member := range members {
// 		parentID := 0
// 		if member.Parent != nil {
// 			parentID = member.Parent.ID
// 		}

// 		jsonNodes = append(jsonNodes, map[string]interface{}{
// 			"ID":            member.ID,
// 			"Position":      member.Position,
// 			"Level":         member.Level,
// 			"PackagePrice":  member.PackagePrice,
// 			"LeftSales":     member.LeftSales,
// 			"RightSales":    member.RightSales,
// 			"SponsorBonus":  member.SponsorBonus,
// 			"BinaryBonus":   member.BinaryBonus,
// 			"MatchingBonus": member.MatchingBonus,
// 			"ParentID":      parentID,
// 			"LeftCarry":     member.LeftCarry,
// 			"RightCarry":    member.RightCarry,
// 		})
// 	}
// 	return jsonNodes
// }

// func sendResultsToDjango(results interface{}) {
// 	jsonData, err := json.Marshal(results)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	resp, err := http.Post("http://localhost:8000/prorevenue := 0.0cess_results/", "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer resp.Body.Close()
// 	fmt.Println("Response from Django:", resp.Status)
// }

func ProcessBinaryTree(data map[string]interface{}) []map[string]interface{} {
	numOfUsers := int(data["num_of_users"].(float64)) + 1
	cycles := int(data["cycle"].(float64))

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

	sponsorBonusPercentage := data["sponsor_bonus_percentage"].(float64)
	binaryBonusPercentage := data["binary_bonus_percentage"].(float64)
	poolBonusPercentage := data["pool_bonus_percentage"].(float64)
	poolBonusCount := data["pool_bonus_count"].(float64)
	matchingBonusPercentages := []float64{}

	if rawPercentages, ok := data["percentage_string"].([]interface{}); ok {
		for _, val := range rawPercentages {
			matchingBonusPercentages = append(matchingBonusPercentages, val.(float64))
		}
	}

	rawCappingScope := data["capping_scope"].([]interface{})
	cappingScope := make([]string, len(rawCappingScope))
	for i, v := range rawCappingScope {
		if str, ok := v.(string); ok {
			cappingScope[i] = str
		}
	}

	ratioChoice := data["ratio_choice"].(string)
	ratioAmount := data["ratio_amount"].(float64)

	cappingAmount := data["capping_amount"].(float64)

	tree := NewTree(numOfUsers, productsPrice)
	tree.Root = &Member{ID: 1, Level: 1}
	tree.Members = append(tree.Members, tree.Root)
	queue := []*Member{tree.Root}

	leftRatioAmount := 0.0
	rightRatioAmount := 0.0

	if binaryBonusPercentage == 0.0 {
		if ratioChoice == "one_one" {
			leftRatioAmount = ratioAmount * 1
			rightRatioAmount = ratioAmount * 1
		} else if ratioChoice == "one_two" {
			leftRatioAmount = ratioAmount * 1
			rightRatioAmount = ratioAmount * 2
		} else if ratioChoice == "two_one" {
			leftRatioAmount = ratioAmount * 2
			rightRatioAmount = ratioAmount * 1
		}
	}

	var totalSponsorBonus = 0.0
	var totalBinaryBonus = 0.0
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
		queue = tree.buildTree(usersPerProduct, queue)
		totalSponsorBonus = tree.setAndGetSponsorBonus(sponsorBonusPercentage, cappingAmount, cappingScope)
		totalBinaryBonus, revenue = tree.setBinaryBonus(cappingAmount, leftRatioAmount, rightRatioAmount, cappingScope, binaryBonusPercentage)
		totalMatchingBonus = tree.setMatchingBonus(matchingBonusPercentages, cappingScope, cappingAmount)
		// treeStructure := convertToJSONStructure(tree.Members)
		adminList := convertToJSONStructureForAdmin(tree.Members)
		adminBinaryBonus := adminList[0]["BinaryBonus"].(float64)
		adminMatchingBonus := adminList[0]["MatchingBonus"].(float64)
		adminSponsorBonus := adminList[0]["SponsorBonus"].(float64)

		totalBinaryBonus = totalBinaryBonus - adminBinaryBonus
		totalMatchingBonus = totalMatchingBonus - adminMatchingBonus
		totalSponsorBonus = totalSponsorBonus - adminSponsorBonus

		expense = totalBinaryBonus + totalSponsorBonus + totalMatchingBonus
		// fmt.Println("Expense:", expense)
		profit = revenue - expense
		if poolBonusPercentage > 0 && poolBonusCount > 0 {
			poolBonus = profit * (poolBonusPercentage / 100)
			profit = profit - poolBonus
			expense += poolBonus
		}

		ans := map[string]interface{}{
			// "tree_structure": convertToJSONStructure(tree.Members),
			// "tree_structure":       adminList,
			"revenue":              revenue,
			"expense":              expense,
			"profit":               profit,
			"pool_bonus":           poolBonus,
			"total_sponsor_bonus":  totalSponsorBonus,
			"total_binary_bonus":   totalBinaryBonus,
			"total_matching_bonus": totalMatchingBonus,
		}
		results = append(results, ans)
	}
	fmt.Println("Data Sending to Django")
	return results
}
