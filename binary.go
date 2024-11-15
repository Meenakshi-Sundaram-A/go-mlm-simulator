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
	Sale                 float64
	SponsorBonus         float64
	BinaryBonus          float64
	LeftSales            float64
	RightSales           float64
	CarryForward         float64
	CarryForwardPosition string
	MatchingBonus        float64
}

type Tree struct {
	Root         *Member
	NumMembers   int
	PackagePrice float64
	Members      []*Member
}

func NewTree(numMembers int, packagePrice float64) *Tree {
	tree := &Tree{
		NumMembers:   numMembers,
		PackagePrice: packagePrice,
	}
	tree.buildTree()
	tree.setMemberSales(packagePrice)
	return tree
}

func (t *Tree) buildTree() {
	if t.NumMembers <= 0 {
		return
	}
	t.Root = &Member{ID: 1, Level: 1}
	t.Members = append(t.Members, t.Root)
	queue := []*Member{t.Root}
	currentID := 2

	for currentID <= t.NumMembers {
		currentMember := queue[0]
		queue = queue[1:]
		if currentID <= t.NumMembers {
			leftChild := &Member{ID: currentID, Parent: currentMember, Position: "Left", Level: currentMember.Level + 1}
			currentMember.LeftMember = leftChild
			queue = append(queue, leftChild)
			t.Members = append(t.Members, leftChild)
			currentID++
		}
		if currentID <= t.NumMembers {
			rightChild := &Member{ID: currentID, Parent: currentMember, Position: "Right", Level: currentMember.Level + 1}
			currentMember.RightMember = rightChild
			queue = append(queue, rightChild)
			t.Members = append(t.Members, rightChild)
			currentID++
		}
	}
}

func (t *Tree) setMemberSales(packagePrice float64) {
	for _, member := range t.Members {
		if member.ID != 1 {
			member.Sale = packagePrice
		}
	}
}

// Calculate Sponsor Bonus
func (t *Tree) setAndGetSponsorBonus(sponsorPercentage, cappingAmount float64, cappingScope string) float64 {
	var totalBonus float64
	for _, member := range t.Members {
		var rightBonus, leftBonus float64
		if member.RightMember != nil {
			rightBonus = member.RightMember.Sale * (sponsorPercentage / 100)
		}
		if member.LeftMember != nil {
			leftBonus = member.LeftMember.Sale * (sponsorPercentage / 100)
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
func (t *Tree) setBinaryBonus(binaryPercentage, cappingAmount float64) float64 {
	totalBonus := 0.0
	for _, member := range t.Members {
		leftSales := 0.0
		rightSales := 0.0
		if member.LeftMember != nil {
			leftSales = t.traverse(member.LeftMember)
			member.LeftSales = leftSales
		}
		if member.RightMember != nil {
			rightSales = t.traverse(member.RightMember)
			member.RightSales = rightSales
		}

		// Calculate binary bonus based on the weaker side
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

	currentSales := node.Sale
	if node.Sale == 0 {
		currentSales = 0
	}

	leftSales := t.traverse(node.LeftMember)
	rightSales := t.traverse(node.RightMember)
	return currentSales + leftSales + rightSales
}

// Calculate Matching Bonus
func (t *Tree) setMatchingBonus(Lev1Percentage float64, Lev2Percentage float64) float64 {
	var totalMatchingBonus float64
	for _, member := range t.Members {

		if member.LeftMember != nil && member.RightMember != nil {
			member.MatchingBonus += member.LeftMember.BinaryBonus * (Lev1Percentage / 100)
			member.MatchingBonus += member.RightMember.BinaryBonus * (Lev1Percentage / 100)
			if member.LeftMember.LeftMember != nil && member.LeftMember.RightMember != nil {
				member.MatchingBonus += member.LeftMember.LeftMember.BinaryBonus * (Lev2Percentage / 100)
				member.MatchingBonus += member.LeftMember.RightMember.BinaryBonus * (Lev2Percentage / 100)
	
			}
			if member.RightMember.LeftMember != nil && member.RightMember.RightMember != nil {
				member.MatchingBonus += member.RightMember.LeftMember.BinaryBonus * (Lev2Percentage / 100)
				member.MatchingBonus += member.RightMember.RightMember.BinaryBonus * (Lev2Percentage / 100)
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
			"ID":                   member.ID,
			"Position":             member.Position,
			"Level":                member.Level,
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
	numOfUsers := int(data["num_of_users"].(float64))
	packagePrice := data["package_price"].(float64)
	sponsorBonusPercentage := data["sponsor_bonus_percentage"].(float64)
	binaryBonusPercentage := data["binary_bonus_percentage"].(float64)
	lev1Percentage := data["lev1_percentage"].(float64)
	lev2Percentage := data["lev2_percentage"].(float64)
	cappingScope := data["capping_scope"].(string)
	cappingAmount := data["capping_amount"].(float64)

	tree := NewTree(numOfUsers, packagePrice)
	sponsorBonus := tree.setAndGetSponsorBonus(sponsorBonusPercentage, cappingAmount, cappingScope)
	totalBinaryBonus := tree.setBinaryBonus(binaryBonusPercentage, cappingAmount)
	totalMatchingBonus := tree.setMatchingBonus(lev1Percentage, lev2Percentage)

	return map[string]interface{}{
		"tree_structure":       convertToJSONStructure(tree.Members),
		"total_sponsor_bonus":  sponsorBonus,
		"total_binary_bonus":   totalBinaryBonus,
		"total_matching_bonus": totalMatchingBonus,
	}
}
