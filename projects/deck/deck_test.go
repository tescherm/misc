package deck

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeck(t *testing.T) {
	t.Parallel()

	deck, err := NewDeck(1)
	require.NoError(t, err)

	require.EqualValues(t, 52, deck.Size())

	card, err := deck.Deal()
	require.NoError(t, err)
	require.Equal(t, KING, card.Value())
	require.Equal(t, DIAMOND, card.Face())

	require.EqualValues(t, 51, deck.Size())

	err = deck.Return(card)
	require.NoError(t, err)

	require.EqualValues(t, 52, deck.Size())

	size := deck.Size()
	cards := make([]Card, size)
	for i := int64(0); i < size; i++ {
		card, err = deck.Deal()
		require.NoError(t, err)
		cards[i] = *card
	}

	require.EqualValues(t, 0, deck.Size())
	card, err = deck.Deal()
	require.Error(t, err)
	require.EqualError(t, err, "deal called on an empty deck")
	require.Nil(t, card)

	for _, card := range cards {
		err = deck.Return(&card)
		require.NoError(t, err)
	}

	require.EqualValues(t, 52, deck.Size())
}

func BenchmarkDeck_Deal(b *testing.B) {
	b.ReportAllocs()

	deck, err := NewDeck(1000)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = deck.Deal()
		require.NoError(b, err)
	}
}

func BenchmarkDeck_Return(b *testing.B) {
	b.ReportAllocs()

	deck, err := NewDeck(1)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		card := NewCard(ACE, SPADE)
		err = deck.Return(&card)
		require.NoError(b, err)
	}
}

func BenchmarkDeck_ShuffleSimple(b *testing.B) {
	b.ReportAllocs()

	deck, err := NewDeck(1000)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = deck.ShuffleSimple()
		require.NoError(b, err)
	}
}

func BenchmarkDeck_ShuffleMMap(b *testing.B) {
	b.ReportAllocs()

	deck, err := NewDeck(1000)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = deck.Shuffle()
		require.NoError(b, err)
	}
}
