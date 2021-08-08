package tunnel

import (
	"context"
	"fmt"
	"io"

	"github.com/OpenSlides/openslides-manage-service/proto"
)

// sendReceiver reads and writes from a grpc tunnel connection.
type sendReceiver interface {
	receiver
	sender
}

type receiver interface {
	Recv() (*proto.TunnelData, error)
}

type sender interface {
	Send(*proto.TunnelData) error
}

// copyStream connects the grcp connection with a io.ReadWriter.
//
// Blocks until one connection is closed.
func copyStream(ctx context.Context, sr sendReceiver, rw io.ReadWriter) error {
	// Create channels to send errors from the goroutines. The channels have to
	// be buffered. If an error in one goroutine happens this function exists.
	// If this happens, the other goroutine still runs. If an error happens, it
	// is wirtten to the channel. An unblocked channel would block forever.
	fromGRPC := make(chan error, 1)
	fromRW := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Message from gRPC
	go func() {
		defer close(fromGRPC)
		if err := receiverToWriter(ctx, rw, sr); err != nil {
			fromGRPC <- fmt.Errorf("copy data from gRPC: %w", err)
		}
	}()

	// Message to gRPC
	go func() {
		defer close(fromRW)

		if err := readerToSender(ctx, sr, rw); err != nil {
			fromRW <- fmt.Errorf("copy data to gRPC: %w", err)
		}
	}()

	// Wait for one side to finish.
	select {
	case err := <-fromGRPC:
		if err != nil {
			return fmt.Errorf("from grpc: %w", err)
		}
	case err := <-fromRW:
		if err != nil {
			return fmt.Errorf("from read writer: %w", err)
		}
	}
	return nil
}

// receiverToWriter copies the data from the receiver (gRPC tunnel) to a writer.
func receiverToWriter(ctx context.Context, w io.Writer, r receiver) error {
	done := make(chan error, 1)
	go func() {
		defer close(done)
		for ctx.Err() != nil {
			c, err := r.Recv()
			if err != nil {
				if err == io.EOF {
					done <- nil
					return
				}
				done <- fmt.Errorf("receiving data: %w", err)
				return
			}

			if _, err := w.Write(c.Data); err != nil {
				done <- fmt.Errorf("writing data: %w", err)
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

//rederToSender copys data from the reader to the sender (gPRC tunnel).
func readerToSender(ctx context.Context, s sender, r io.Reader) error {
	done := make(chan error, 1)
	go func() {
		defer close(done)

		buff := make([]byte, 1<<20) // 1 MB buffer
		for ctx.Err() != nil {
			n, err := r.Read(buff)
			if err != nil {
				if err == io.EOF {
					done <- nil
					return
				}
				done <- fmt.Errorf("receiving data: %w", err)
				return
			}

			if err := s.Send(&proto.TunnelData{Data: buff[:n]}); err != nil {
				done <- fmt.Errorf("writing data: %w", err)
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
