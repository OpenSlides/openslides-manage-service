package manage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/OpenSlides/openslides-manage-service/pkg/datastore"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
)

const (
	authHashPath      = "/internal/auth/hash"
	datastorWritePath = "/internal/datastore/writer/write"
)

const setPasswordHelp = `Sets the password of an user

This command sets the password of a user by a given user id.
`

// CmdSetPassword initializes the set-password command.
func CmdSetPassword(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-password",
		Short: "Sets an user password.",
		Long:  setPasswordHelp,
	}

	userID := cmd.Flags().Int64P("user_id", "u", 1, "ID of the user account.")
	password := cmd.Flags().StringP("password", "p", "admin", "New password for the user.")
	onlyUpdate := cmd.Flags().Bool("update", false, "Do not overwrite an existing password.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		service, close, err := Dial(ctx, cfg.Address)
		if err != nil {
			return fmt.Errorf("connecting to gRPC server: %w", err)
		}
		defer close()

		req := &proto.SetPasswordRequest{
			UserID:     *userID,
			Password:   *password,
			OnlyUpdate: *onlyUpdate,
		}

		resp, err := service.SetPassword(ctx, req)
		if err != nil {
			return fmt.Errorf("reset password: %w", err)
		}

		if *onlyUpdate {
			msg := "Password did already exist."
			if resp.PasswordSet {
				msg = "Password was set."
			}
			fmt.Println(msg)
		}
		return nil
	}

	return cmd
}

// SetPassword sets hashes and sets the password.
func (s *Server) SetPassword(ctx context.Context, in *proto.SetPasswordRequest) (*proto.SetPasswordResponse, error) {
	waitForService(ctx, s.config.AuthURL().Host, s.config.DatastoreWriterURL().Host)

	if in.OnlyUpdate {
		// Check if the current password exists.
		waitForService(ctx, s.config.DatastoreReaderURL().Host)
		key := fmt.Sprintf("user/%d/password", in.UserID)
		var oldPassword string
		if err := datastore.Get(ctx, s.config.DatastoreReaderURL().String(), key, &oldPassword); err != nil {
			return nil, fmt.Errorf("fetching old password: %w", err)
		}

		if oldPassword != "" {
			return &proto.SetPasswordResponse{PasswordSet: false}, nil
		}
	}

	hash, err := hashPassword(ctx, s.config.AuthURL(), in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := setPassword(ctx, s.config.DatastoreWriterURL(), int(in.UserID), hash); err != nil {
		return nil, fmt.Errorf("set password: %w", err)
	}
	return &proto.SetPasswordResponse{PasswordSet: true}, nil
}

// hashPassword returns the hashed form of password as a JSON.
func hashPassword(ctx context.Context, authAddr *url.URL, password string) (string, error) {
	reqBody := fmt.Sprintf(`{"toHash": "%s"}`, password)
	authAddr.Path = authHashPath
	req, err := http.NewRequestWithContext(ctx, "POST", authAddr.String(), strings.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("creating request to auth service: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("[can not read body]")
		}
		return "", fmt.Errorf("auth service returned %s: %s", resp.Status, body)
	}

	var respBody struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	return respBody.Hash, nil
}

func setPassword(ctx context.Context, writerURL *url.URL, userID int, hash string) error {
	key := fmt.Sprintf("user/%d/password", userID)
	value := []byte(`"` + hash + `"`)
	if err := datastore.Set(ctx, writerURL.String(), key, value); err != nil {
		return fmt.Errorf("writing key %s to %s: %w", key, writerURL.String(), err)
	}

	return nil
}
