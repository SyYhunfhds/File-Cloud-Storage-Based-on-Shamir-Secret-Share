package ui

import (
	"fmt"
	"image/color"

	"simulator/internal/algorithm"
	"simulator/internal/business"
	"simulator/internal/visualization"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whiteColor = color.RGBA{255, 255, 255, 255}
)

type SceneType int

const (
	SceneMainMenu SceneType = iota
	SceneRoleManagement
	SceneProposalManagement
	SceneApproval
	SceneSettings
)

type Button struct {
	x      float64
	y      float64
	width  float64
	height float64
	text   string
	action func()
}

type AnimationState struct {
	animationType    string
	progress     float64
	totalTime    float64
	completed    bool
	shareX       float64
	shareY       float64
	targetX      float64
	targetY      float64
	proposalID   string
}

type Game struct {
	config           *algorithm.Config
	shamir           *algorithm.Shamir
	roleManager      *business.RoleManager
	proposalManager  *business.ProposalManager
	shareManager     *business.ShareManager
	canvas           *visualization.Canvas
	currentScene     SceneType
	screenWidth      float64
	screenHeight     float64
	animationTime    float64
	buttons          []Button
	mouseX           float64
	mouseY           float64
	mousePressed     bool
	lastMousePressed bool
	hoveredButtonIdx int
	message          string
	messageTime      float64
	animation        *AnimationState
}

func NewGame(config *algorithm.Config) (*Game, error) {
	shamir, err := algorithm.NewShamir(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create shamir: %w", err)
	}

	roleManager := business.NewRoleManager(shamir)
	proposalManager := business.NewProposalManager(shamir)
	shareManager := business.NewShareManager(shamir)
	canvas := visualization.NewCanvas(400, 300)

	return &Game{
		config:           config,
		shamir:          shamir,
		roleManager:      roleManager,
		proposalManager:  proposalManager,
		shareManager:     shareManager,
		canvas:           canvas,
		currentScene:     SceneMainMenu,
		screenWidth:      1280,
		screenHeight:     720,
		animationTime:    0,
		buttons:          make([]Button, 0),
		mouseX:          0,
		mouseY:          0,
		mousePressed:    false,
		lastMousePressed: false,
		hoveredButtonIdx: -1,
		message:         "",
		messageTime:     0,
		animation:        nil,
	}, nil
}

func (g *Game) Update() error {
	g.animationTime += 1.0 / 60.0

	if g.messageTime > 0 {
		g.messageTime -= 1.0 / 60.0
		if g.messageTime <= 0 {
			g.message = ""
		}
	}

	if g.animation != nil {
		if !g.animation.completed {
			g.animation.progress += 1.0 / 60.0
			if g.animation.progress >= g.animation.totalTime {
				g.animation.progress = g.animation.totalTime
				g.animation.completed = true
			}
		} else {
			g.animation = nil
		}
	}

	var cursorX, cursorY int
	cursorX, cursorY = ebiten.CursorPosition()
	g.mouseX, g.mouseY = float64(cursorX), float64(cursorY)
	g.mousePressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	g.hoveredButtonIdx = -1
	for i, btn := range g.buttons {
		if g.isPointInRect(g.mouseX, g.mouseY, btn.x, btn.y, btn.width, btn.height) {
			g.hoveredButtonIdx = i
			if g.mousePressed && !g.lastMousePressed && g.animation == nil {
				btn.action()
			}
			break
		}
	}

	g.lastMousePressed = g.mousePressed

	return nil
}

func (g *Game) isPointInRect(px, py, rx, ry, rw, rh float64) bool {
	return px >= rx && px <= rx+rw && py >= ry && py <= ry+rh
}

func (g *Game) showMessage(msg string) {
	g.message = msg
	g.messageTime = 3.0
}

func (g *Game) startShareAnimation(proposalID string, share algorithm.Share) {
	g.animation = &AnimationState{
		animationType:    "share",
		progress:   0,
		totalTime:  1.0,
		completed:  false,
		shareX:     float64(share.X),
		shareY:     float64(share.Y),
		targetX:    5.0,
		targetY:    0.0,
		proposalID: proposalID,
	}
}

func (g *Game) startRecoveryAnimation(proposalID string) {
	g.animation = &AnimationState{
		animationType:    "recovery",
		progress:   0,
		totalTime:  1.5,
		completed:  false,
		proposalID: proposalID,
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 46, 255})

	ebitenutil.DrawRect(screen, 0, 0, 1280, 60, color.RGBA{70, 130, 180, 255})
	g.drawText(screen, "Shamir Secret Sharing Simulator", 30, 20, whiteColor)

	ebitenutil.DrawRect(screen, 0, 60, 200, 660, color.RGBA{45, 45, 65, 255})

	g.buttons = g.buttons[:0]

	g.drawButton(screen, 20, 80, 160, 40, "Role Management", func() {
		g.currentScene = SceneRoleManagement
	})
	g.drawButton(screen, 20, 130, 160, 40, "Proposal Management", func() {
		g.currentScene = SceneProposalManagement
	})
	g.drawButton(screen, 20, 180, 160, 40, "Approval Process", func() {
		g.currentScene = SceneApproval
	})
	g.drawButton(screen, 20, 230, 160, 40, "System Settings", func() {
		g.currentScene = SceneSettings
	})

	ebitenutil.DrawRect(screen, 200, 60, 880, 660, color.RGBA{50, 50, 75, 255})
	ebitenutil.DrawRect(screen, 1080, 60, 200, 660, color.RGBA{45, 45, 65, 255})

	switch g.currentScene {
	case SceneMainMenu:
		g.drawMainMenu(screen)
	case SceneRoleManagement:
		g.drawRoleManagement(screen)
	case SceneProposalManagement:
		g.drawProposalManagement(screen)
	case SceneApproval:
		g.drawApproval(screen)
	case SceneSettings:
		g.drawSettings(screen)
	}

	ebitenutil.DrawRect(screen, 1080, 60, 200, 660, color.RGBA{45, 45, 65, 255})
	g.drawText(screen, "Visualization", 1100, 80, whiteColor)

	if g.currentScene == SceneApproval {
		g.drawVisualization(screen)
	}

	ebitenutil.DrawRect(screen, 0, 680, 1280, 40, color.RGBA{35, 35, 55, 255})
	g.drawText(screen, fmt.Sprintf("Threshold: %d  |  Shares: %d  |  Prime: %d",
		g.shamir.GetThreshold(), g.shamir.GetNumShares(), g.shamir.GetPrime()), 20, 690, whiteColor)
	g.drawText(screen, "Version: 1.0.0", 1100, 690, whiteColor)

	if g.message != "" {
		ebitenutil.DrawRect(screen, 400, 350, 480, 40, color.RGBA{0, 100, 0, 255})
		g.drawText(screen, g.message, 420, 360, whiteColor)
	}
}

func (g *Game) drawText(screen *ebiten.Image, text string, x, y int, col color.RGBA) {
	ebitenutil.DebugPrintAt(screen, text, x, y)
}

func (g *Game) drawButton(screen *ebiten.Image, x, y, w, h float64, text string, action func()) {
	btnIdx := len(g.buttons)
	g.buttons = append(g.buttons, Button{x: x, y: y, width: w, height: h, text: text, action: action})

	var btnColor color.RGBA
	if g.hoveredButtonIdx == btnIdx {
		btnColor = color.RGBA{100, 149, 237, 255}
	} else {
		btnColor = color.RGBA{70, 130, 180, 255}
	}

	ebitenutil.DrawRect(screen, x, y, w, h, btnColor)
	g.drawText(screen, text, int(x+10), int(y+10), whiteColor)
}

func (g *Game) drawMainMenu(screen *ebiten.Image) {
	g.drawText(screen, "Welcome to Shamir Secret Sharing Simulator", 250, 100, whiteColor)
	g.drawText(screen, "Features:", 250, 140, whiteColor)
	g.drawText(screen, "1. Role Management - Add, edit, delete approval roles", 270, 170, whiteColor)
	g.drawText(screen, "2. Proposal Management - Create fund transfer proposals", 270, 200, whiteColor)
	g.drawText(screen, "3. Approval Process - Collect shares and recover key", 270, 230, whiteColor)
	g.drawText(screen, "4. System Settings - Configure threshold and parameters", 270, 260, whiteColor)
	g.drawText(screen, "Click buttons on the left to switch functions", 250, 320, whiteColor)
	g.drawText(screen, "Press ESC to exit the program", 250, 350, whiteColor)
}

func (g *Game) drawRoleManagement(screen *ebiten.Image) {
	g.drawText(screen, "=== Role Management ===", 250, 100, whiteColor)

	roles := g.roleManager.ListRoles()
	g.drawText(screen, fmt.Sprintf("Current roles: %d", len(roles)), 250, 140, whiteColor)

	y := 180
	for i, role := range roles {
		g.drawText(screen, fmt.Sprintf("%d. %s (Weight: %d)", i+1, role.Name, role.Weight), 270, y, whiteColor)
		y += 30
	}

	g.drawButton(screen, 250, 500, 120, 40, "Add Role", func() {
		g.roleManager.AddRole(fmt.Sprintf("Role%d", len(g.roleManager.ListRoles())+1), 1)
		g.showMessage("New role added")
	})

	g.drawButton(screen, 390, 500, 120, 40, "Distribute Shares", func() {
		currentRoles := g.roleManager.ListRoles()
		if len(currentRoles) > 0 {
			g.roleManager.DistributeShares(42)
			g.showMessage("Shares distributed")
		} else {
			g.showMessage("Please add roles first")
		}
	})

	g.drawButton(screen, 530, 500, 120, 40, "Back to Menu", func() {
		g.currentScene = SceneMainMenu
	})
}

func (g *Game) drawProposalManagement(screen *ebiten.Image) {
	g.drawText(screen, "=== Proposal Management ===", 250, 100, whiteColor)

	proposals := g.proposalManager.ListProposals()
	g.drawText(screen, fmt.Sprintf("Current proposals: %d", len(proposals)), 250, 140, whiteColor)

	y := 180
	for i, p := range proposals {
		g.drawText(screen, fmt.Sprintf("%d. %s - %s (Amount: %d)", i+1, p.Title, p.Status.String(), p.Amount), 270, y, whiteColor)
		y += 30
	}

	g.drawButton(screen, 250, 500, 120, 40, "Create Proposal", func() {
		g.proposalManager.CreateProposal("Fund Transfer Request", "Test proposal", int64(10000))
		g.showMessage("New proposal created")
	})

	g.drawButton(screen, 390, 500, 120, 40, "Back to Menu", func() {
		g.currentScene = SceneMainMenu
	})
}

func (g *Game) drawApproval(screen *ebiten.Image) {
	g.drawText(screen, "=== Approval Process ===", 250, 100, whiteColor)

	pendingProposals := g.proposalManager.ListPendingProposals()
	g.drawText(screen, fmt.Sprintf("Pending proposals: %d", len(pendingProposals)), 250, 140, whiteColor)

	if len(pendingProposals) > 0 {
		p := pendingProposals[0]
		g.drawText(screen, fmt.Sprintf("Current proposal: %s", p.Title), 270, 180, whiteColor)
		g.drawText(screen, fmt.Sprintf("Shares needed: %d/%d", p.Submitted, p.Required), 270, 210, whiteColor)
	} else {
		g.drawText(screen, "No pending proposals", 270, 180, whiteColor)
	}

	g.drawButton(screen, 250, 500, 120, 40, "Submit Share", func() {
		pendingProposals := g.proposalManager.ListPendingProposals()
		if len(pendingProposals) > 0 {
			proposal := pendingProposals[0]
			share := algorithm.Share{X: int64(proposal.Submitted + 1), Y: 100}
			g.shareManager.SubmitShare(proposal.ID, share)
			g.proposalManager.SubmitShare(proposal.ID)
			g.startShareAnimation(proposal.ID, share)
			g.showMessage("Share submitted")
		} else {
			g.showMessage("Please create proposal first")
		}
	})

	g.drawButton(screen, 390, 500, 120, 40, "Recover Key", func() {
		pendingProposals := g.proposalManager.ListPendingProposals()
		if len(pendingProposals) > 0 && g.shareManager.CanRecover(pendingProposals[0].ID) {
			g.startRecoveryAnimation(pendingProposals[0].ID)
			secret, _ := g.shareManager.RecoverSecret(pendingProposals[0].ID)
			g.showMessage(fmt.Sprintf("Recovered key: %d", secret))
		} else if len(pendingProposals) > 0 {
			g.showMessage("Not enough shares to recover")
		} else {
			g.showMessage("Please create proposal first")
		}
	})

	g.drawButton(screen, 530, 500, 120, 40, "Back to Menu", func() {
		g.currentScene = SceneMainMenu
	})
}

func (g *Game) drawVisualization(screen *ebiten.Image) {
	visualizationRectX := 1100.0
	visualizationRectY := 100.0
	visualizationRectWidth := 180.0
	visualizationRectHeight := 560.0

	ebitenutil.DrawRect(screen, visualizationRectX, visualizationRectY, visualizationRectWidth, visualizationRectHeight, color.RGBA{60, 60, 85, 255})

	pendingProposals := g.proposalManager.ListPendingProposals()
	if len(pendingProposals) == 0 {
		g.drawText(screen, "No proposal selected", int(visualizationRectX+10), int(visualizationRectY+20), whiteColor)
		return
	}

	proposal := pendingProposals[0]
	shares := g.shareManager.GetSubmittedShares(proposal.ID)

	g.canvas.Clear()
	g.canvas.SetScale(5.0)
	g.canvas.SetOffset(visualizationRectX+visualizationRectWidth/2, visualizationRectY+visualizationRectHeight/2)

	for _, share := range shares {
		g.canvas.AddPoint(float64(share.X), float64(share.Y))
	}

	if g.animation != nil && !g.animation.completed {
		switch g.animation.animationType {
		case "share":
			t := g.animation.progress / g.animation.totalTime
			x := g.animation.shareX*(1-t) + g.animation.targetX*t
			y := g.animation.shareY*(1-t) + g.animation.targetY*t
			g.canvas.AddPoint(x, y)
		case "recovery":
			t := g.animation.progress / g.animation.totalTime
			if t > 0.5 {
				g.drawRecoveryAnimation(screen, visualizationRectX, visualizationRectY, visualizationRectWidth, visualizationRectHeight, t-0.5)
			}
		}
	}

	g.canvas.Draw(screen)

	if len(shares) > 0 {
		g.drawText(screen, fmt.Sprintf("Shares: %d", len(shares)), int(visualizationRectX+10), int(visualizationRectY+530), whiteColor)
	}
}

func (g *Game) drawRecoveryAnimation(screen *ebiten.Image, x, y, width, height float64, progress float64) {
	centerX := x + width/2
	centerY := y + height/2
	radius := 50.0 * progress

	vector.StrokeCircle(screen, float32(centerX), float32(centerY), float32(radius), 2, color.RGBA{100, 255, 100, 255}, true)

	if progress >= 1.0 {
		g.drawText(screen, "Key Recovered!", int(centerX-40), int(centerY-10), whiteColor)
	}
}

func (g *Game) drawSettings(screen *ebiten.Image) {
	g.drawText(screen, "=== System Settings ===", 250, 100, whiteColor)
	g.drawText(screen, "Current configuration:", 250, 140, whiteColor)
	g.drawText(screen, fmt.Sprintf("Threshold: %d", g.shamir.GetThreshold()), 270, 170, whiteColor)
	g.drawText(screen, fmt.Sprintf("Shares: %d", g.shamir.GetNumShares()), 270, 200, whiteColor)
	g.drawText(screen, fmt.Sprintf("Prime: %d", g.shamir.GetPrime()), 270, 230, whiteColor)

	g.drawText(screen, "System Information:", 250, 280, whiteColor)
	g.drawText(screen, fmt.Sprintf("Roles: %d", g.roleManager.Count()), 270, 310, whiteColor)
	g.drawText(screen, fmt.Sprintf("Proposals: %d", g.proposalManager.Count()), 270, 340, whiteColor)

	g.drawButton(screen, 250, 400, 120, 40, "Reset Shares", func() {
		g.showMessage("Shares reset")
	})

	g.drawButton(screen, 390, 400, 120, 40, "Clear Proposals", func() {
		g.showMessage("Proposals cleared")
	})

	g.drawButton(screen, 530, 400, 120, 40, "Back to Menu", func() {
		g.currentScene = SceneMainMenu
	})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(g.screenWidth), int(g.screenHeight)
}

func (g *Game) GetRoleManager() *business.RoleManager {
	return g.roleManager
}

func (g *Game) GetProposalManager() *business.ProposalManager {
	return g.proposalManager
}

func (g *Game) GetShareManager() *business.ShareManager {
	return g.shareManager
}

func (g *Game) GetShamir() *algorithm.Shamir {
	return g.shamir
}

func (g *Game) GetCanvas() *visualization.Canvas {
	return g.canvas
}

func (g *Game) GetAnimationTime() float64 {
	return g.animationTime
}

func (g *Game) SetScene(scene SceneType) {
	g.currentScene = scene
}

func (g *Game) GetScene() SceneType {
	return g.currentScene
}
