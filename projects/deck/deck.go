package deck

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/edsrzf/mmap-go"
)

const cardOffset = 32

type Card byte

func NewCard(value Value, face Face) Card {
	return Card(int(face)*cardOffset + int(value))
}

func (c *Card) Face() Face {
	return Face(int(*c) / cardOffset)
}

func (c *Card) Value() Value {
	return Value(int(*c) % cardOffset)
}

func (c *Card) String() string {
	return fmt.Sprintf("%s of %ss", c.Value(), c.Face())
}

type Face int
type Value int

const (
	ACE Value = iota
	TWO
	THREE
	FOUR
	FIVE
	SIX
	SEVEN
	EIGHT
	NINE
	TEN
	JACK
	QUEEN
	KING
)

func (v Value) String() string {
	values := [...]string{"Ace", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Jack", "Queen", "King"}
	if len(values) < int(v) {
		return ""
	}

	return values[v]
}

var Values = []Value{
	ACE,
	TWO,
	THREE,
	FOUR,
	FIVE,
	SIX,
	SEVEN,
	EIGHT,
	NINE,
	TEN,
	JACK,
	QUEEN,
	KING,
}

const (
	SPADE Face = iota
	HEART
	CLUB
	DIAMOND
)

func (f Face) String() string {
	faces := [...]string{"Spade", "Heart", "Club", "Diamond"}
	if len(faces) < int(f) {
		return ""
	}

	return faces[f]
}

var Faces = []Face{
	SPADE,
	HEART,
	CLUB,
	DIAMOND,
}

type Deck struct {
	sync.RWMutex

	topIdx int64

	file  *os.File
	cards mmap.MMap
	rnd   *rand.Rand
}

func NewDeck(iterations int64) (*Deck, error) {
	f, err := os.CreateTemp("", "deck")
	if err != nil {
		return nil, fmt.Errorf("create deck file failed: %w", err)
	}

	deckSize := iterations * int64(len(Faces)) * int64(len(Values))
	err = f.Truncate(deckSize)
	if err != nil {
		return nil, fmt.Errorf("deck truncate failed: %w", err)
	}

	cards, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("deck mmap failed: %w", err)
	}

	currentSize := int64(0)
	for i := int64(0); i < iterations; i++ {
		for _, face := range Faces {
			for _, value := range Values {
				card := NewCard(value, face)
				cards[currentSize] = byte(card)
				currentSize++
			}
		}
	}

	err = cards.Flush()
	if err != nil {
		return nil, fmt.Errorf("deck flush failed: %w", err)
	}

	deck := &Deck{
		topIdx: currentSize - 1,
		file:   f,
		cards:  cards,

		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return deck, nil
}

func (d *Deck) Size() int64 {
	d.RLock()
	defer d.RUnlock()

	return d.topIdx + 1
}

func (d *Deck) Close() error {
	err := d.cards.Flush()
	if err != nil {
		return fmt.Errorf("deck file %s flush failed: %w", d.file.Name(), err)
	}
	err = d.cards.Unmap()
	if err != nil {
		return fmt.Errorf("deck file %s unmap failed: %w", d.file.Name(), err)
	}
	err = d.file.Close()
	if err != nil {
		return fmt.Errorf("deck file %s close failed: %w", d.file.Name(), err)
	}
	return nil
}

func (d *Deck) Print() error {
	d.RLock()
	defer d.RUnlock()

	for i := int64(0); i < d.topIdx; i++ {
		card := Card(d.cards[i])
		fmt.Printf("%s\n", card.String())
	}

	return nil
}

func (d *Deck) Deal() (*Card, error) {
	d.Lock()
	defer d.Unlock()

	if d.topIdx < 0 {
		return nil, fmt.Errorf("deal called on an empty deck")
	}

	card := Card(d.cards[d.topIdx])
	d.topIdx--
	return &card, nil
}

func (d *Deck) Return(c *Card) error {
	d.Lock()
	defer d.Unlock()

	if c == nil {
		return fmt.Errorf("cannot return a nil card")
	}

	d.topIdx++
	d.cards[d.topIdx] = byte(*c)

	return nil
}

func (d *Deck) Shuffle() error {
	d.Lock()
	defer d.Unlock()

	fisherYatesShuffle(d.cards, d.topIdx, d.rnd)

	return nil
}

func ShuffleSimple(filename string) error {
	cards, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read deck file %s failed: %w", filename, err)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	fisherYatesShuffle(cards, int64(len(cards)), rnd)

	err = ioutil.WriteFile(filename, cards, 0644)
	if err != nil {
		return fmt.Errorf("write deck file %s failed: %w", filename, err)
	}

	return nil
}

func fisherYatesShuffle(arr []byte, n int64, rnd *rand.Rand) {
	for i := n - 1; i >= 0; i-- {
		j := int64(rnd.Intn(int(n-i))) + i
		tmp := arr[i]
		arr[i] = arr[j]
		arr[j] = tmp
	}
}
