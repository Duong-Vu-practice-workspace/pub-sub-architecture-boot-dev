package __subscriber

func (u user) doBattles(subCh <-chan move) []piece {
	// ?
	var battles []piece
	for move := range subCh {
		for _, p := range u.pieces {
			if p.location == move.piece.location {
				battles = append(battles, p)
			}
		}
	}
	return battles
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

func (u user) march(p piece, publishCh chan<- move) {
	publishCh <- move{
		userName: u.name,
		piece:    p,
	}
}

func distributeBattles(publishCh <-chan move, subChans []chan move) {
	for mv := range publishCh {
		for _, subCh := range subChans {
			subCh <- mv
		}
	}
}
