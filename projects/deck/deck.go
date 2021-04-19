package deck

import (
	"bufio"
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

	size int64
	file string
	rnd  *rand.Rand
}

func NewDeck(iterations int64) (*Deck, error) {
	f, err := os.CreateTemp("", "deck")
	if err != nil {
		return nil, fmt.Errorf("create deck file failed")
	}

	defer f.Close()

	w := bufio.NewWriter(f)
	for i := int64(0); i < iterations; i++ {
		for _, face := range Faces {
			for _, value := range Values {
				card := NewCard(value, face)
				err = w.WriteByte(byte(card))
				if err != nil {
					return nil, fmt.Errorf("error writing card: %w", err)
				}
			}
		}
	}

	err = w.Flush()
	if err != nil {
		return nil, fmt.Errorf("deck flush failed: %w", err)
	}

	deck := &Deck{
		size: iterations * int64(len(Faces)) * int64(len(Values)),
		file: f.Name(),

		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return deck, nil
}

func (d *Deck) Size() int64 {
	d.RLock()
	defer d.RUnlock()

	return d.size
}

func (d *Deck) Print() error {
	d.RLock()
	defer d.RUnlock()

	f, err := os.OpenFile(d.file, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("open deck file %s failed: %w", d.file, err)
	}

	reader := bufio.NewReader(f)

	for i := int64(0); i < d.size; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("read card failed: %w", err)
		}

		card := Card(b)
		fmt.Printf("%s\n", card.String())
	}

	return nil
}

func (d *Deck) Deal() (*Card, error) {
	d.Lock()
	defer d.Unlock()

	f, err := os.OpenFile(d.file, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open deck file %s failed: %w", d.file, err)
	}

	if d.size == 0 {
		return nil, fmt.Errorf("deal called on an empty deck")
	}

	buf := make([]byte, 1)
	_, err = f.ReadAt(buf, d.size-1)
	if err != nil {
		return nil, fmt.Errorf("read card failed: %w", err)
	}

	d.size--
	err = f.Truncate(d.size)
	if err != nil {
		return nil, fmt.Errorf("deck file %s truncate failed: %w", d.file, err)
	}

	card := Card(buf[0])
	return &card, nil
}

func (d *Deck) Return(c *Card) error {
	d.Lock()
	defer d.Unlock()

	if c == nil {
		return fmt.Errorf("cannot return a nil card")
	}

	f, err := os.OpenFile(d.file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open deck file %s failed: %w", d.file, err)
	}

	defer f.Close()

	_, err = f.Write([]byte{byte(*c)})
	if err != nil {
		return fmt.Errorf("card return failed: %w", err)
	}

	d.size++

	return nil
}

func (d *Deck) Shuffle() error {
	d.Lock()
	defer d.Unlock()

	f, err := os.OpenFile(d.file, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open deck file %s failed: %w", d.file, err)
	}

	defer f.Close()
	cards, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		return fmt.Errorf("open deck file %s failed: %w", d.file, err)
	}

	defer cards.Unmap()

	d.fisherYatesShuffle(cards)

	err = cards.Flush()
	if err != nil {
		return fmt.Errorf("mmap flush failed: %w", err)
	}

	return nil
}

func (d *Deck) ShuffleSimple() error {
	d.Lock()
	defer d.Unlock()

	cards, err := ioutil.ReadFile(d.file)
	if err != nil {
		return fmt.Errorf("read deck file %s failed: %w", d.file, err)
	}

	d.fisherYatesShuffle(cards)

	err = ioutil.WriteFile(d.file, cards, 0644)
	if err != nil {
		return fmt.Errorf("write deck file %s failed: %w", d.file, err)
	}

	return nil
}

func (d *Deck) fisherYatesShuffle(arr []byte) {
	for i := len(arr) - 1; i >= 0; i-- {
		j := d.rnd.Intn(len(arr)-i) + i
		tmp := arr[i]
		arr[i] = arr[j]
		arr[j] = tmp
	}
}
