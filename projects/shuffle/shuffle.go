package shuffle

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

type pile struct {
	f *os.File
	w *bufio.Writer
}

type Shuffler struct {
	in  *os.File
	out *os.File
	rnd *rand.Rand

	piles []pile
}

func NewShuffler(inFile, outFile string) (*Shuffler, error) {
	in, err := createInputFile(inFile)
	if err != nil {
		return nil, err
	}

	out, err := os.OpenFile(outFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("create dest in failed: %w", err)
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// selected such that len(file) / m fits into RAM
	m := 6
	piles := make([]pile, m)

	s := &Shuffler{
		in:  in,
		out: out,
		rnd: rnd,

		piles: piles,
	}

	return s, nil
}

func createInputFile(name string) (*os.File, error) {
	in, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("create source in failed: %w", err)
	}

	inWriter := bufio.NewWriter(in)

	// 1GB / 8 bytes per int
	for i := 0; i < 1e+9/8; i++ {
		_, err := inWriter.WriteString(fmt.Sprintf("%d\n", i))
		if err != nil {
			return nil, fmt.Errorf("source in write failed: %w", err)
		}
	}

	err = inWriter.Flush()
	if err != nil {
		return nil, fmt.Errorf("source in flush failed: %w", err)
	}

	_, err = in.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("source file seek failed: %w", err)
	}

	return in, nil
}

func (s *Shuffler) FirstPass() error {
	for i := 0; i < len(s.piles); i++ {
		f, err := os.CreateTemp("", "shuffled")
		if err != nil {
			return fmt.Errorf("create shuffle in failed: %pileWriter", err)
		}
		pileWriter := bufio.NewWriter(f)
		s.piles[i] = pile{
			f: f,
			w: pileWriter,
		}
	}

	// map input value to a random pile
	scanner := bufio.NewScanner(s.in)
	for scanner.Scan() {
		pile := s.piles[s.rnd.Intn(len(s.piles))]
		_, err := pile.w.WriteString(scanner.Text())
		if err != nil {
			return fmt.Errorf("shuffle file write failed: %w", err)
		}

		_, err = pile.w.WriteRune('\n')
		if err != nil {
			return fmt.Errorf("shuffle file write failed: %w", err)
		}
	}

	for _, pile := range s.piles {
		err := pile.w.Flush()
		if err != nil {
			return fmt.Errorf("shuffle file flush failed: %w", err)
		}

		_, err = pile.f.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("shuffle file seek failed: %w", err)
		}
	}

	return nil
}

func (s *Shuffler) SecondPass() error {
	outWriter := bufio.NewWriter(s.out)

	for i := 0; i < len(s.piles); i++ {
		pile := s.piles[i]
		err := s.write(pile, outWriter)
		if err != nil {
			return err
		}
	}

	err := outWriter.Flush()
	if err != nil {
		return fmt.Errorf("output file flush failed: %w", err)
	}

	return nil
}

func (s *Shuffler) write(p pile, w *bufio.Writer) error {
	scanner := bufio.NewScanner(p.f)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	s.rnd.Shuffle(len(lines), func(i, j int) {
		lines[i], lines[j] = lines[j], lines[i]
	})

	for _, line := range lines {
		_, err := w.WriteString(line)
		if err != nil {
			return fmt.Errorf("output file write failed: %w", err)
		}

		_, err = w.WriteRune('\n')
		if err != nil {
			return fmt.Errorf("output file write failed: %w", err)
		}
	}

	return nil
}

func (s *Shuffler) Close() {
	s.in.Close()
	s.out.Close()

	for _, pile := range s.piles {
		pile.f.Close()
		os.Remove(pile.f.Name())
	}
}
