// binary.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Member struct {
	ID                   int
	Parent               *Member
	LeftMember           *Member
	RightMember          *Member
	Position             string
	Level                int
	PackagePrice         float64
	Sale                 float64
	SponsorBonus         float64
	BinaryBonus          float64
	MatchingBonus        float64
	LeftSales            float64
	RightSales           float64
	CarryForward         float64
	CarryForwardPosition string
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

	//tree.setMemberSales(productsPrice)
	return tree
}

func sumSlice(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

func (t *Tree) buildTree(productsPrice []float64, usersPerProduct []float64, queue []*Member) []*Member {
	currCount := 0
	if t.NumMembers <= 0 {
		return queue
	}
	totalUsersPerCycle := sumSlice(usersPerProduct)
	// if len(queue) == 0 {
	// 	return queue // Avoid accessing an empty queue
	// }
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

// func (t *Tree) buildTree() {
// 	if t.NumMembers <= 0 {
// 		return
// 	}
// 	t.Root = &Member{ID: 1, Level: 1}
// 	t.Members = append(t.Members, t.Root)
// 	queue := []*Member{t.Root}
// 	currentID := 2

// 	for currentID <= t.NumMembers {
// 		currentMember := queue[0]
// 		queue = queue[1:]
// 		if currentID <= t.NumMembers {
// 			leftChild := &Member{ID: currentID, Parent: currentMember, Position: "Left", Level: currentMember.Level + 1}
// 			currentMember.LeftMember = leftChild
// 			queue = append(queue, leftChild)
// 			t.Members = append(t.Members, leftChild)
// 			currentID++
// 		}
// 		if currentID <= t.NumMembers {
// 			rightChild := &Member{ID: currentID, Parent: currentMember, Position: "Right", Level: currentMember.Level + 1}
// 			currentMember.RightMember = rightChild
// 			queue = append(queue, rightChild)
// 			t.Members = append(t.Members, rightChild)
// 			currentID++
// 		}
// 	}
// }

// func (t *Tree) setMemberSales(packagePrice float64) {
// 	for _, member := range t.Members {
// 		if member.ID != 1 {
// 			member.Sale = packagePrice
// 		}
// 	}
// }

func (t *Tree) setAndGetSponsorBonus(sponsorPercentage, cappingAmount float64, cappingScope string) float64 {
	var totalBonus float64
	for _, member := range t.Members {
		var rightBonus, leftBonus float64
		if member.RightMember != nil {
			rightBonus = member.RightMember.PackagePrice * (sponsorPercentage / 100)
		}
		if member.LeftMember != nil {
			leftBonus = member.LeftMember.PackagePrice * (sponsorPercentage / 100)
		}
		sponsorBonus := rightBonus + leftBonus
		if cappingAmount > 0 && cappingScope == "3" && sponsorBonus > cappingAmount {
			member.SponsorBonus += cappingAmount
		} else {
			member.SponsorBonus += sponsorBonus
		}
		totalBonus += member.SponsorBonus
	}
	return totalBonus
}

// Calculate Binary Bonus
func (t *Tree) setBinaryBonus(binaryPercentage, cappingAmount float64) float64 {
	totalBonus := 0.0
	for _, member := range t.Members {
		leftSales := 0.0
		rightSales := 0.0
		if member.LeftMember != nil {
			leftSales = t.traverse(member.LeftMember)
			if member.CarryForward > 0.0 && member.CarryForwardPosition == "Left" {
				leftSales += member.CarryForward
				member.CarryForward = 0.0
				member.CarryForwardPosition = ""
			}
			member.LeftSales = leftSales
		}
		if member.RightMember != nil {
			rightSales = t.traverse(member.RightMember)
			if member.CarryForward > 0.0 && member.CarryForwardPosition == "Right" {
				rightSales += member.CarryForward
				member.CarryForward = 0.0
				member.CarryForwardPosition = ""
			}
			member.RightSales = rightSales
		}

		fmt.Print("Node:", member.ID, "Left Sale:", member.LeftSales, "Right Sale:", member.RightSales)

		weakerSideSales := leftSales
		if rightSales < leftSales {
			weakerSideSales = rightSales
		}

		binaryBonus := weakerSideSales * (binaryPercentage / 100)
		if cappingAmount > 0 && binaryBonus > cappingAmount {
			member.BinaryBonus = cappingAmount
		} else {
			member.BinaryBonus = binaryBonus
		}

		carryForward := leftSales - rightSales
		if member.LeftMember != nil && carryForward > 0 {
			member.CarryForward = carryForward
			member.CarryForwardPosition = "Left"
		} else if member.RightMember != nil && carryForward < 0 {
			member.CarryForward = -carryForward
			member.CarryForwardPosition = "Right"
		}

		totalBonus += member.BinaryBonus
	}
	return totalBonus
}

func (t *Tree) traverse(node *Member) float64 {
	if node == nil {
		return 0
	}

	currentSales := node.PackagePrice
	// if node.Sale == 0 {
	// 	currentSales = 0
	// }

	leftSales := t.traverse(node.LeftMember)
	rightSales := t.traverse(node.RightMember)
	return currentSales + leftSales + rightSales
}

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
		print(member.MatchingBonus)
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
			"ID":                   member.ID,
			"Position":             member.Position,
			"Level":                member.Level,
			"PackagePrice":         member.PackagePrice,
			"LeftSales":            member.LeftSales,
			"RightSales":           member.RightSales,
			"SponsorBonus":         member.SponsorBonus,
			"BinaryBonus":          member.BinaryBonus,
			"MatchingBonus":        member.MatchingBonus,
			"ParentID":             parentID,
			"CarryForward":         member.CarryForward,
			"CarryForwardPosition": member.CarryForwardPosition,
		})
	}
	return jsonNodes
}

func (t *Tree) DisplayTree() {
	queue := []*Member{t.Root}
	for len(queue) > 0 {
		currentMember := queue[0]
		queue = queue[1:]
		// fmt.Printf("Member ID: %d, Sponsor Bonus: %.2f, Binary Bonus: %.2f, Matching Bonus: %.2f\n",
		// 	currentMember.ID, currentMember.SponsorBonus, currentMember.BinaryBonus, currentMember.MatchingBonus)

		if currentMember.LeftMember != nil {
			queue = append(queue, currentMember.LeftMember)
		}
		if currentMember.RightMember != nil {
			queue = append(queue, currentMember.RightMember)
		}
	}
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

func ProcessBinaryTree(data map[string]interface{}) map[string]interface{} {
	numOfUsers := int(data["num_of_users"].(float64)) + 1
	//packagePrice := data["package_price"].(float64)
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
	matchingBonusPercentages := []float64{}

	if rawPercentages, ok := data["percentage_string"].([]interface{}); ok {
		for _, val := range rawPercentages {
			matchingBonusPercentages = append(matchingBonusPercentages, val.(float64))
		}
	}

	cappingScope := data["capping_scope"].(string)
	cappingAmount := data["capping_amount"].(float64)

	tree := NewTree(numOfUsers, productsPrice)
	tree.Root = &Member{ID: 1, Level: 1}
	tree.Members = append(tree.Members, tree.Root)
	queue := []*Member{tree.Root}

	var sponsorBonus = 0.0
	for i := 0; i < cycles; i++ {
		usersPerProduct := []float64{}
		if rawPercentages, ok := data["users_per_product"].([]interface{}); ok {
			for _, val := range rawPercentages {
				usersPerProduct = append(usersPerProduct, val.(float64))
			}
		}
		queue = tree.buildTree(productsPrice, usersPerProduct, queue)
		sponsorBonus += tree.setAndGetSponsorBonus(sponsorBonusPercentage, cappingAmount, cappingScope)
	}

	totalBinaryBonus := tree.setBinaryBonus(binaryBonusPercentage, cappingAmount)
	totalMatchingBonus := tree.setMatchingBonus(matchingBonusPercentages)

	return map[string]interface{}{
		"tree_structure":       convertToJSONStructure(tree.Members),
		"total_sponsor_bonus":  sponsorBonus,
		"total_binary_bonus":   totalBinaryBonus,
		"total_matching_bonus": totalMatchingBonus,
	}
}
