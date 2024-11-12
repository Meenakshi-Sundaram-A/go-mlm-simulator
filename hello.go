package main

import (
	
	"fmt"
	
)

type FormData struct {
    NumOfUsers             int     `json:"num_of_users"`
    PackagePrice           float64 `json:"package_price"`
    SponsorBonusPercentage int     `json:"sponsor_bonus_percentage"`
    BinaryBonusPercentage  int     `json:"binary_bonus_percentage"`
    Lev1Percentage         int     `json:"lev1_percentage"`
    Lev2Percentage         int     `json:"lev2_percentage"`
    CappingScope           string  `json:"capping_scope"`
    CappingAmount          int     `json:"capping_amount"`
    CarryYesNo             string  `json:"carry_yes_no"`
}

type Member struct {
	ID            int
	Parent        *Member
	LeftMember    *Member
	RightMember   *Member
	Position      string
	Level         int
	Sale          float64
	SponsorBonus  float64
	BinaryBonus   float64
	LeftSales     float64
	RightSales    float64
	CarryForward  float64
	MatchingBonus float64
}

type Tree struct {
	Root                   *Member
	NumMembers             int
	PackagePrice           float64
	AdditionalProductPrice float64
	Members                []*Member
}

func NewTree(numMembers int, packagePrice, additionalProductPrice float64) *Tree {
	tree := &Tree{
		NumMembers:             numMembers,
		PackagePrice:           packagePrice,
		AdditionalProductPrice: additionalProductPrice,
	}
	tree.buildTree()
	tree.setMemberSales(packagePrice, additionalProductPrice)
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

func (t *Tree) setMemberSales(packagePrice, additionalProductPrice float64) {
	for _, member := range t.Members {
		if member.ID != 1 {
			member.Sale = packagePrice + additionalProductPrice
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

		// Calculate carry forward
		carryForward := leftSales - rightSales
		if member.LeftMember != nil && carryForward > 0 {
			member.LeftMember.CarryForward = carryForward
		} else if member.RightMember != nil && carryForward < 0 {
			member.RightMember.CarryForward = -carryForward
		}

		// Add the member's binary bonus to the total bonus
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
func (t *Tree) setMatchingBonus(matchingPercentage float64) float64 {
	var totalMatchingBonus float64
	for _, member := range t.Members {
		if member.LeftMember != nil {
			member.MatchingBonus += member.LeftMember.BinaryBonus * (matchingPercentage / 100)
		}
		if member.RightMember != nil {
			member.MatchingBonus += member.RightMember.BinaryBonus * (matchingPercentage / 100)
		}
		totalMatchingBonus += member.MatchingBonus
	}
	return totalMatchingBonus
}

func (t *Tree) DisplayTree() {
	queue := []*Member{t.Root}
	for len(queue) > 0 {
		currentMember := queue[0]
		queue = queue[1:]
		fmt.Printf("Member ID: %d, Sponsor Bonus: %.2f, Binary Bonus: %.2f, Matching Bonus: %.2f\n",
			currentMember.ID, currentMember.SponsorBonus, currentMember.BinaryBonus, currentMember.MatchingBonus)

		if currentMember.LeftMember != nil {
			queue = append(queue, currentMember.LeftMember)
		}
		if currentMember.RightMember != nil {
			queue = append(queue, currentMember.RightMember)
		}
	}
}

func main() {
	numMembers := 1000000
	packagePrice := 1000.0
	additionalProductPrice := 0.0
	sponsorPercentage := 10.0
	cappingAmount := 0.0
	cappingScope := "3"
	binaryPercentage := 10.0
	matchingPercentage := 2.0

	tree := NewTree(numMembers, packagePrice, additionalProductPrice)
	sponsorBonus := tree.setAndGetSponsorBonus(sponsorPercentage, cappingAmount, cappingScope)
	totalBinaryBonus := tree.setBinaryBonus(binaryPercentage, cappingAmount)
	totalMatchingBonus := tree.setMatchingBonus(matchingPercentage)

	tree.DisplayTree()
	fmt.Printf("Total Sponsor Bonus: %.2f\n", sponsorBonus)
	fmt.Printf("Total Binary Bonus: %.2f\n", totalBinaryBonus)
	fmt.Printf("Total Matching Bonus: %.2f\n", totalMatchingBonus)
}
