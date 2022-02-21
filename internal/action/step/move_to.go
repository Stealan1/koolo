package step

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type MoveToStep struct {
	basicStep
	toX      int
	toY      int
	teleport bool
	path     []astar.Pather
}

func MoveTo(toX, toY int, teleport bool) *MoveToStep {
	return &MoveToStep{
		basicStep: newBasicStep(),
		toX:       toX,
		toY:       toY,
		teleport:  teleport,
	}
}

func (m *MoveToStep) Status(data game.Data) Status {
	_, distance, _ := pather.GetPathToDestination(data, m.toX, m.toY)
	if distance < 8 {
		return m.tryTransitionStatus(StatusCompleted)
	}

	return m.status
}

func (m *MoveToStep) Run(data game.Data) error {
	if m.teleport && m.status == StatusNotStarted {
		hid.PressKey(config.Config.Bindings.Teleport)
	}
	m.tryTransitionStatus(StatusInProgress)

	// TODO: In case of teleport, calculate fcr frames for waiting time
	if time.Since(m.lastRun) < time.Millisecond*500 {
		return nil
	}

	if m.path == nil || !m.adjustPath(data) {
		// TODO: Handle not found
		path, _, _ := pather.GetPathToDestination(data, m.toX, m.toY)
		m.path = path
	}

	m.lastRun = time.Now()
	pather.MoveThroughPath(m.path, 25, m.teleport)

	return nil
}

// Cache the path and try to reuse it
func (m *MoveToStep) adjustPath(data game.Data) bool {
	nearestKey := 0
	nearestDistance := 99999999
	for k, pos := range m.path {
		distance := pather.DistanceFromPoint(data, pos.(*pather.Tile).X+data.AreaOrigin.X, pos.(*pather.Tile).Y+data.AreaOrigin.Y)
		if distance < nearestDistance {
			nearestDistance = distance
			nearestKey = k
		}
	}

	if nearestDistance < 5 && len(m.path) > nearestKey {
		//fmt.Println(fmt.Sprintf("Max deviation: %d, using Path Key: %d [%d]", nearestDistance, nearestKey, len(m.path)-1))
		m.path = m.path[:nearestKey]

		return true
	}

	return false
}
