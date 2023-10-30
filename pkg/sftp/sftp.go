package sftp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/fatih/color"
	"github.com/melbahja/goph"

	"github.com/relex/aini"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

func UploadFileAndExecute(client *goph.Client, localFilePath, remoteFilePath string) error {
	// Check if the local file exists
	_, err := os.Stat(localFilePath)
	if err != nil {
		return fmt.Errorf("error checking local file: %w", err)
	}

	// Upload the local file to the remote server
	fmt.Printf("Uploading file %s to %s\n", localFilePath, remoteFilePath)
	err = client.Upload(localFilePath, remoteFilePath)
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}
	fmt.Println("File uploaded successfully")

	// Make the remote file executable
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

func ExecuteCommandOnHost(sftpFile, file, host string) error {
	//file = flag.String("i", "", "Ansible inventory file")
	//host = flag.String("h", "", "Host to connect to")
	//flag.Parse()

	// Parse flags
	// Parse the inventory file

	// Parse the inventory file
	inv, err := aini.ParseFile(file)
	if err != nil {
		return fmt.Errorf("error parsing inventory file: %w", err)
	}

	// Find the host in the inventory
	h, ok := inv.Hosts[host]
	if !ok {
		return fmt.Errorf("error: host %s not found in inventory", host)
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
