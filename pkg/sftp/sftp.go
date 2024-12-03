package sftp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/fatih/color"
	"github.com/melbahja/goph"

	"github.com/relex/aini"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

func UploadFileAndExecute(client *goph.Client, localFilePath, remoteFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	defer func() {
		if _, err := client.Run("rm -f " + remoteFilePath); err != nil {
			fmt.Printf("Warning: failed to cleanup remote file: %v\n", err)
		}
	}()

	if _, err := os.Stat(localFilePath); err != nil {
		return fmt.Errorf("error accessing local file: %w", err)
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("upload timed out")
	default:
		if err := client.Upload(localFilePath, remoteFilePath); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
	}

	fmt.Printf("Uploading file %s to %s\n", localFilePath, remoteFilePath)
	err := client.Upload(localFilePath, remoteFilePath)
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}
	fmt.Println("File uploaded successfully")

	fmt.Printf("Making file %s executable\n", remoteFilePath)
	_, err = client.Run("chmod +x " + remoteFilePath)
	if err != nil {
		return fmt.Errorf("error making file executable: %w", err)
	}
	fmt.Println("File made executable successfully")

	/*
		// Execute the remote file
		fmt.Printf("Executing file %s\n", remoteFilePath)
		output, err := client.Run(remoteFilePath)
		if err != nil {
			return fmt.Errorf("error executing file: %w", err)
		}
		fmt.Println("File executed successfully")
		fmt.Println("Output:")
		fmt.Println(output)
	*/
	return nil
}

func ExecuteCommandOnHost(sftpFile, inventoryFile, hostName string) error {
	//file = flag.String("i", "", "Ansible inventory file")
	//host = flag.String("h", "", "Host to connect to")
	//flag.Parse()

	// Parse flags
	// Parse the inventory file

	// Parse the inventory file
	inv, err := aini.ParseFile(inventoryFile)
	if err != nil {
		return fmt.Errorf("error parsing inventory file: %w", err)
	}

	// Find the host in the inventory
	h, ok := inv.Hosts[hostName]
	if !ok {
		return fmt.Errorf("error: host %s not found in inventory", hostName)
	}

	// Execute the command on the specified host
	auth, err := goph.UseAgent()
	if err != nil {
		return fmt.Errorf("error using SSH agent: %w", err)
	}

	client, err := goph.New(strings.TrimPrefix(h.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(h.Vars["ansible_host"], "Vars: "), auth)
	if err != nil {
		return fmt.Errorf("error creating SSH client: %w", err)
	}

	// Extract the filename from the local file path
	localFileName := filepath.Base(sftpFile)

	// Define the remote directory path
	remoteDir := "/tmp/"

	// Combine the remote directory path with the local file name to get the remote file path
	remoteFilePath := filepath.Join(remoteDir, localFileName)

	// Upload the file and execute it
	err = UploadFileAndExecute(client, sftpFile, remoteFilePath)
	if err != nil {
		return fmt.Errorf("error uploading and executing file: %w", err)
	}

	return nil
}
