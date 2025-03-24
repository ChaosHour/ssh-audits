package sftp

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/fatih/color"
	"github.com/melbahja/goph"
	"github.com/relex/aini"
	"golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

// Add this helper function to the sftp package
func AddNewHost(hostname string, remote net.Addr, key ssh.PublicKey) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	knownHostsFile := filepath.Join(home, ".ssh", "known_hosts")

	// Check if known_hosts file exists and contains the host
	if hostData, err := os.ReadFile(knownHostsFile); err == nil {
		if strings.Contains(string(hostData), string(ssh.MarshalAuthorizedKey(key))) {
			fmt.Printf("[*] Host %s already in known_hosts\n", hostname)
			return nil
		}
	}

	// Add the new host to known hosts file
	err = goph.AddKnownHost(hostname, remote, key, knownHostsFile)
	if err != nil {
		return fmt.Errorf("failed to add known host: %w", err)
	}

	fmt.Printf("[+] Added %s to known hosts\n", hostname)
	return nil
}

func UploadFileAndExecute(client *goph.Client, localFilePath, remoteFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if _, err := os.Stat(localFilePath); err != nil {
		return fmt.Errorf("error accessing local file: %w", err)
	}

	// Upload the file
	fmt.Printf("Uploading file %s to %s\n", localFilePath, remoteFilePath)
	select {
	case <-ctx.Done():
		return fmt.Errorf("upload timed out")
	default:
		if err := client.Upload(localFilePath, remoteFilePath); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
	}
	fmt.Println("File uploaded successfully")

	// Make executable
	fmt.Printf("Making file %s executable\n", remoteFilePath)
	_, err := client.Run("chmod +x " + remoteFilePath)
	if err != nil {
		// Cleanup on error
		client.Run("rm -f " + remoteFilePath)
		return fmt.Errorf("error making file executable: %w", err)
	}
	fmt.Println("File made executable successfully")

	// Execute file
	fmt.Printf("Executing file %s\n", remoteFilePath)
	output, err := client.Run(remoteFilePath)
	if err != nil {
		// Cleanup on error
		client.Run("rm -f " + remoteFilePath)
		return fmt.Errorf("error executing file: %w\nOutput: %s", err, string(output))
	}
	fmt.Println("File executed successfully")
	fmt.Println("Output:")
	// Convert output bytes to string and trim any whitespace
	fmt.Println(strings.TrimSpace(string(output)))

	// Cleanup after successful execution
	if _, err := client.Run("rm -f " + remoteFilePath); err != nil {
		fmt.Printf("Warning: failed to cleanup remote file: %v\n", err)
	}

	return nil
}

// Add new function for direct connections
func ExecuteCommandOnHostDirect(sftpFile, hostname string) error {
	auth, err := goph.UseAgent()
	if err != nil {
		return fmt.Errorf("error using SSH agent: %w", err)
	}

	config := &goph.Config{
		Auth: auth,
		User: os.Getenv("USER"),
		Addr: hostname,
		Port: uint(22),
		Callback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return AddNewHost(hostname, remote, key)
		},
	}

	client, err := goph.NewConn(config)
	if err != nil {
		return fmt.Errorf("error creating SSH client: %w", err)
	}
	defer client.Close()

	localFileName := filepath.Base(sftpFile)
	remoteFilePath := filepath.Join("/tmp/", localFileName)

	return UploadFileAndExecute(client, sftpFile, remoteFilePath)
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
