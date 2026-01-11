package __publisher

func (u user) march(p piece, publishCh chan<- move) {
	// ?
	publishCh <- move{
		userName: u.name,
		piece:    p,
	}
}

// don't touch below this line

type user struct {
	name   string
	pieces []piece
}

type move struct {
	userName string
	piece    piece
}

type piece struct {
	location string
	name     string
}

func doBattles(publishCh <-chan move, users []user) []piece {
	fights := []piece{}
	for mv := range publishCh {
		for _, u := range users {
			if u.name == mv.userName {
				continue
			}
			for _, piece := range u.pieces {
				if piece.location == mv.piece.location {
					fights = append(fights, piece)
				}
			}
		}
	}
	return fights
}
