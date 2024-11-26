// binary.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
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

func (t *Tree) setAndGetSponsorBonus(sponsorPercentage, cappingAmount float64, cappingScope string) float64 {
	totalBonus := 0.0
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
		if cappingAmount > 0 && cappingScope == "3" && sponsorBonus > cappingAmount {
			member.SponsorBonus = cappingAmount
		} else {
			member.SponsorBonus = sponsorBonus
		}
		totalBonus += member.SponsorBonus
	}
	return totalBonus
}

// Calculate Binary Bonus
// func (t *Tree) setBinaryBonus(cappingAmount float64, leftRatioAmount float64, rightRatioAmount float64) float64 {
// 	totalBonus := 0.0
// 	for _, member := range t.Members {
// 		leftSales := 0.0
// 		rightSales := 0.0
// 		if member.LeftMember != nil {
// 			leftSales = t.traverse(member.LeftMember)
// 			if member.LeftCarry > 0.0 {
// 				leftSales += member.LeftCarry
// 				member.LeftCarry = 0.0
// 			}
// 			member.LeftSales = leftSales
// 		}
// 		if member.RightMember != nil {
// 			rightSales = t.traverse(member.RightMember)
// 			if member.RightCarry > 0.0 {
// 				rightSales += member.RightCarry
// 				member.RightCarry = 0.0
// 			}
// 			member.RightSales = rightSales
// 		}

// 		//fmt.Println("Node:", member.ID, "Left Sale:", member.LeftSales, "Right Sale:", member.RightSales)

// 		pairCount := int(math.Min((leftSales / leftRatioAmount), (rightSales / rightRatioAmount)))
// 		leftVal := float64(pairCount) * leftRatioAmount
// 		rightVal := float64(pairCount) * rightRatioAmount
// 		minValue := math.Min(leftVal, rightVal)

// 		if pairCount <= 5 {
// 			binaryBonus := minValue * (50.0 / 100)
// 			if cappingAmount > 0 && binaryBonus > cappingAmount {
// 				member.BinaryBonus = cappingAmount
// 			} else {
// 				member.BinaryBonus = binaryBonus
// 			}
// 		} else if pairCount > 5 && pairCount <= 10 {
// 			binaryBonus := minValue * (25.0 / 100)
// 			if cappingAmount > 0 && binaryBonus > cappingAmount {
// 				member.BinaryBonus = cappingAmount
// 			} else {
// 				member.BinaryBonus = binaryBonus
// 			}
// 		} else if pairCount > 10 {
// 			binaryBonus := minValue * (25.0 / 100)
// 			if cappingAmount > 0 && binaryBonus > cappingAmount {
// 				member.BinaryBonus = cappingAmount
// 			} else {
// 				member.BinaryBonus = binaryBonus
// 			}
// 		}
// 		member.LeftCarry = leftSales - (float64(pairCount) * leftRatioAmount)
// 		member.RightCarry = rightSales - (float64(pairCount) * rightRatioAmount)

// 		//fmt.Println("Node:", member.ID, "Binary:", member.BinaryBonus, "Left Sale:", member.LeftSales, "Right Sale:", member.RightSales, "Left Carry:", member.LeftCarry, "Right Carry:", member.RightCarry)
// 		totalBonus += member.BinaryBonus
// 	}
// 	return totalBonus
// }

func (t *Tree) setBinaryBonus(cappingAmount float64, leftRatioAmount float64, rightRatioAmount float64) float64 {
	totalBonus := 0.0
	cache := make(map[*Member]float64) // Create cache map

	for _, member := range t.Members {
		leftSales := 0.0
		rightSales := 0.0

		if member.LeftMember != nil {
			leftSales = t.traverse(member.LeftMember, cache)
			if member.LeftCarry > 0.0 {
				leftSales += member.LeftCarry
				member.LeftCarry = 0.0
			}
			member.LeftSales = leftSales
		}

		if member.RightMember != nil {
			rightSales = t.traverse(member.RightMember, cache)
			if member.RightCarry > 0.0 {
				rightSales += member.RightCarry
				member.RightCarry = 0.0
			}
			member.RightSales = rightSales
		}

		pairCount := int(math.Min((leftSales / leftRatioAmount), (rightSales / rightRatioAmount)))
		leftVal := float64(pairCount) * leftRatioAmount
		rightVal := float64(pairCount) * rightRatioAmount
		minValue := math.Min(leftVal, rightVal)

		if pairCount <= 5 {
			binaryBonus := minValue * (50.0 / 100)
			if cappingAmount > 0 && binaryBonus > cappingAmount {
				member.BinaryBonus = cappingAmount
			} else {
				member.BinaryBonus = binaryBonus
			}
		} else if pairCount > 5 && pairCount <= 10 {
			binaryBonus := minValue * (25.0 / 100)
			if cappingAmount > 0 && binaryBonus > cappingAmount {
				member.BinaryBonus = cappingAmount
			} else {
				member.BinaryBonus = binaryBonus
			}
		} else if pairCount > 10 {
			binaryBonus := minValue * (25.0 / 100)
			if cappingAmount > 0 && binaryBonus > cappingAmount {
				member.BinaryBonus = cappingAmount
			} else {
				member.BinaryBonus = binaryBonus
			}
		}
		member.LeftCarry = leftSales - (float64(pairCount) * leftRatioAmount)
		member.RightCarry = rightSales - (float64(pairCount) * rightRatioAmount)

		//fmt.Println("Node:", member.ID, "Binary:", member.BinaryBonus, "Left Sale:", member.LeftSales, "Right Sale:", member.RightSales, "Left Carry:", member.LeftCarry, "Right Carry:", member.RightCarry)
		// totalBonus += member.BinaryBonus

		totalBonus += member.BinaryBonus
	}

	return totalBonus
}

func (t *Tree) traverse(node *Member, cache map[*Member]float64) float64 {
	if node == nil {
		return 0
	}

	// Check if the result for this node is already computed
	if val, exists := cache[node]; exists {
		return val
	}

	// Calculate the current sales recursively
	currentSales := node.PackagePrice
	leftSales := t.traverse(node.LeftMember, cache)
	rightSales := t.traverse(node.RightMember, cache)

	totalSales := currentSales + leftSales + rightSales

	// Store the result in the cache
	cache[node] = totalSales

	return totalSales
}

// func (t *Tree) traverse(node *Member) float64 {
// 	if node == nil {
// 		return 0
// 	}

// 	currentSales := node.PackagePrice

// 	leftSales := t.traverse(node.LeftMember)
// 	rightSales := t.traverse(node.RightMember)
// 	return currentSales + leftSales + rightSales
// }

func (t *Tree) setMatchingBonus(levelPercentages []float64) float64 {
	totalMatchingBonus := 0.0

	for _, member := range t.Members {
		member.MatchingBonus = 0.0
		queue := []*Member{member}

		for _, percentage := range levelPercentages {
			nextLevelNodes := []*Member{}

			for _, node := range queue {
				if node.LeftMember != nil {
					member.MatchingBonus += node.LeftMember.BinaryBonus * (percentage / 100)
					nextLevelNodes = append(nextLevelNodes, node.LeftMember)
				}
				if node.RightMember != nil {
					member.MatchingBonus += node.RightMember.BinaryBonus * (percentage / 100)
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

func convertToJSONStructure(members []*Member) []map[string]interface{} {
	var jsonNodes []map[string]interface{}
	for _, member := range members {
		parentID := 0
		if member.Parent != nil {
			parentID = member.Parent.ID
		}

		jsonNodes = append(jsonNodes, map[string]interface{}{
			"ID":            member.ID,
			"Position":      member.Position,
			"Level":         member.Level,
			"PackagePrice":  member.PackagePrice,
			"LeftSales":     member.LeftSales,
			"RightSales":    member.RightSales,
			"SponsorBonus":  member.SponsorBonus,
			"BinaryBonus":   member.BinaryBonus,
			"MatchingBonus": member.MatchingBonus,
			"ParentID":      parentID,
			"LeftCarry":     member.LeftCarry,
			"RightCarry":    member.RightCarry,
		})
	}
	return jsonNodes
}

func sendResultsToDjango(results interface{}) {
	jsonData, err := json.Marshal(results)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post("http://localhost:8000/process_results/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	fmt.Println("Response from Django:", resp.Status)
}

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
	matchingBonusPercentages := []float64{}

	if rawPercentages, ok := data["percentage_string"].([]interface{}); ok {
		for _, val := range rawPercentages {
			matchingBonusPercentages = append(matchingBonusPercentages, val.(float64))
		}
	}
	ratioChoice := data["ratio_choice"].(string)
	ratioAmount := data["ratio_amount"].(float64)
	cappingScope := data["capping_scope"].(string)
	cappingAmount := data["capping_amount"].(float64)

	tree := NewTree(numOfUsers, productsPrice)
	tree.Root = &Member{ID: 1, Level: 1}
	tree.Members = append(tree.Members, tree.Root)
	queue := []*Member{tree.Root}

	leftRatioAmount := 0.0
	rightRatioAmount := 0.0

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

	var sponsorBonus = 0.0
	var totalBinaryBonus = 0.0
	var totalMatchingBonus = 0.0
	var results []map[string]interface{}
	for i := 0; i < cycles; i++ {
		usersPerProduct := []float64{}
		if rawPercentages, ok := data["users_per_product"].([]interface{}); ok {
			for _, val := range rawPercentages {
				usersPerProduct = append(usersPerProduct, val.(float64))
			}
		}
		queue = tree.buildTree(usersPerProduct, queue)
		sponsorBonus = tree.setAndGetSponsorBonus(sponsorBonusPercentage, cappingAmount, cappingScope)
		totalBinaryBonus = tree.setBinaryBonus(cappingAmount, leftRatioAmount, rightRatioAmount)
		totalMatchingBonus = tree.setMatchingBonus(matchingBonusPercentages)
		// treeStructure := convertToJSONStructure(tree.Members)

		ans := map[string]interface{}{
			"tree_structure":       convertToJSONStructure(tree.Members),
			"total_sponsor_bonus":  sponsorBonus,
			"total_binary_bonus":   totalBinaryBonus,
			"total_matching_bonus": totalMatchingBonus,
		}
		results = append(results, ans)
	}
	// for i, value := range results {
	// 	fmt.Print(i, " ", value)
	// }
	fmt.Println("Data Sending to Django")
	return results
}
