package stew

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marwanhawari/stew/constants"
)

// LockFile contains all the data for the lockfile
type LockFile struct {
	Os       string        `json:"os"`
	Arch     string        `json:"arch"`
	Packages []PackageData `json:"packages"`
}

// PackageData contains the information for an installed binary
type PackageData struct {
	Source     string `json:"source"`
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	Tag        string `json:"tag"`
	Asset      string `json:"asset"`
	Binary     string `json:"binary"`
	URL        string `json:"url"`
	BinaryHash string `json:"binaryHash"`
}

func ReadLockFileJSON(lockFilePath string) (LockFile, error) {

	lockFileBytes, err := os.ReadFile(lockFilePath)
	if err != nil {
		return LockFile{}, err
	}

	var lockFile LockFile
	err = json.Unmarshal(lockFileBytes, &lockFile)
	if err != nil {
		return LockFile{}, err
	}

	return lockFile, nil
}

// WriteLockFileJSON will write the lockfile JSON file
func WriteLockFileJSON(lockFileJSON LockFile, outputPath string) error {

	lockFileBytes, err := json.MarshalIndent(lockFileJSON, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, lockFileBytes, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("📄 Updated %v\n", constants.GreenColor(outputPath))

	return nil
}

// RemovePackage will remove a package from a LockFile.Packages slice
func RemovePackage(pkgs []PackageData, index int) ([]PackageData, error) {
	if len(pkgs) == 0 {
		return []PackageData{}, NoPackagesInLockfileError{}
	}

	if index < 0 || index >= len(pkgs) {
		return []PackageData{}, IndexOutOfBoundsInLockfileError{}
	}

	return append(pkgs[:index], pkgs[index+1:]...), nil
}

// ReadStewfileContents will read the contents of the Stewfile
func ReadStewfileContents(stewfilePath string) ([]PackageData, error) {
	file, err := os.Open(stewfilePath)
	if err != nil {
		return []PackageData{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var packages []PackageData
	for scanner.Scan() {
		packageText := scanner.Text()
		pkg, err := ParseCLIInput(packageText)
		if err != nil {
			return []PackageData{}, err
		}
		packages = append(packages, pkg)
	}

	if err := scanner.Err(); err != nil {
		return []PackageData{}, err
	}

	return packages, nil
}

// ReadStewLockFileContents will read the contents of the Stewfile.lock.json
func ReadStewLockFileContents(lockFilePath string) ([]PackageData, error) {
	lockFile, err := ReadLockFileJSON(lockFilePath)
	if err != nil {
		return []PackageData{}, err
	}
	return lockFile.Packages, nil
}

// NewLockFile creates a new instance of the LockFile struct
func NewLockFile(stewLockFilePath, userOS, userArch string) (LockFile, error) {
	var lockFile LockFile
	lockFileExists, err := PathExists(stewLockFilePath)
	if err != nil {
		return LockFile{}, err
	}
	if !lockFileExists {
		lockFile = LockFile{Os: userOS, Arch: userArch, Packages: []PackageData{}}
	} else {
		lockFile, err = ReadLockFileJSON(stewLockFilePath)
		if err != nil {
			return LockFile{}, err
		}
	}
	return lockFile, nil
}

// DeleteAssetAndBinary will delete the asset from the ~/.stew/pkg path and delete the binary from the ~/.stew/bin path
func DeleteAssetAndBinary(stewPkgPath, stewBinPath, asset, binary string) error {
	assetPath := filepath.Join(stewPkgPath, asset)
	binPath := filepath.Join(stewBinPath, binary)
	err := os.RemoveAll(assetPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(binPath)
	if err != nil {
		return err
	}
	return nil
}

func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
