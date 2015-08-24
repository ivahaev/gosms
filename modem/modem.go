package modem

import (
	"fmt"
	log "github.com/ivahaev/go-logger"
	"github.com/ivahaev/gosms/pdu"
	"github.com/tarm/serial"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	tpmr = 0
)

type GSMModem struct {
	ComPort  string
	BaudRate int
	Port     *serial.Port
	DeviceId string
}

func New(ComPort string, BaudRate int, DeviceId string) (modem *GSMModem) {
	modem = &GSMModem{ComPort: ComPort, BaudRate: BaudRate, DeviceId: DeviceId}
	return modem
}

func (m *GSMModem) Connect() (err error) {
	config := &serial.Config{Name: m.ComPort, Baud: m.BaudRate, ReadTimeout: time.Second}
	m.Port, err = serial.OpenPort(config)
	return err
}

func (m *GSMModem) SendCommand(command string, waitForOk bool) string {
	log.Info("SendCommand: ", command)
	var status string = ""
	m.Port.Flush()
	_, err := m.Port.Write([]byte(command))
	if err != nil {
		log.Error(err)
		return ""
	}
	buf := make([]byte, 32)
	var loop int = 1
	if waitForOk {
		loop = 10
	}
	for i := 0; i < loop; i++ {
		// ignoring error as EOF raises error on Linux
		n, _ := m.Port.Read(buf)
		if n > 0 {
			status = string(buf[:n])
			log.Info("SendCommand: rcvd bytes: ", n, status)
			if strings.Contains(status, "OK\r\n") || strings.Contains(status, "ERROR\r\n") {
//			if strings.HasSuffix(status, "OK\r\n") || strings.HasSuffix(status, "ERROR\r\n") {
				break
			}
		}
	}
	return status
}

func (m *GSMModem) SendSMS(mobile string, message string) string {
	log.Info("SendSMS ", mobile, message)
	mobile = strings.Replace(mobile, "+", "", -1)
	// detected a double-width char
	if len([]rune(message)) < len(message) {
		log.Info("This is UNICODE sms. Will use PDU mode")
		return m.SendPduSMS(mobile, message)
	}
	// Put Modem in SMS Text Mode
	m.SendCommand("AT+CMGF=1\r", false)

	m.SendCommand("AT+CMGS=\""+mobile+"\"\r", false)

	// EOM CTRL-Z = 26
	return m.SendCommand(message+string(26), true)

}

func (m *GSMModem) SendPduSMS(mobile string, message string) string {
	log.Info("SendPduSMS ", mobile, message)
	if len([]rune(message)) > 70 {
		log.Info("This is long message. Will split")
		return m.SendLongPduSms(mobile, message)
	}
	// Put Modem in SMS Binary Mode
	status := m.SendCommand("AT+CMGF=0\r", false)

	telNumber := "01" + "00" + fmt.Sprintf("%02X", len(mobile)) + "91" + encodePhoneNumber(mobile)
	encodedText := pdu.EncodeUcs2ToString(message)
	textLen := lenInHex(encodedText)
	text := telNumber + "0008" + textLen + encodedText

	status = m.SendCommand("AT+CMGS="+strconv.Itoa(lenInBytes(text))+"\r", false)
	text = "00" + text
	// EOM CTRL-Z = 26
	status = m.SendCommand(text+string(26), true)
	log.Notice("Message status:", status)
	return status

}

func (m *GSMModem) SendLongPduSms(mobile string, message string) string {
	mes := []rune(message)
	numberOfMessages := len(mes) / 67
	if len(mes) % 67 > 0 {
		numberOfMessages++
	}
	log.Debug("Total messages", numberOfMessages, "length:", len(message))
	udh := createUDH(numberOfMessages)
	encodedPhoneNumber := encodePhoneNumber(mobile)
	phoneLength := fmt.Sprintf("%02X", len(mobile))
	var status string
	for i := 0; i < numberOfMessages; i++ {
		status = m.SendCommand("AT+CMGF=0\r", false)
		log.Debug(status)
		telNumber := "41" + getNextTpmr() + phoneLength + "91" + encodedPhoneNumber
		startByte := i * 67
		stopByte := (i + 1) * 67
		if stopByte >= len(mes) {
			stopByte = len(mes) - 1
		}
		text := string(mes[startByte:stopByte])
		log.Debug(startByte, stopByte, text)
		encodedText := pdu.EncodeUcs2ToString(text)
		textLen := lenInHex(encodedText)
		text = telNumber + "0008" + udh[i] + textLen + encodedText
		status = m.SendCommand("AT+CMGS="+strconv.Itoa(lenInBytes(text))+"\r", false)
		log.Debug(status)
		text = "00" + text
		status = m.SendCommand(text+string(26), true)
		log.Debug(status)
	}
	return status
}

func lenInHex(str string) string {
	return fmt.Sprintf("%02X", lenInBytes(str))
}

func lenInBytes(str string) int {
	return int(float64(len(str))/2 + 0.9999)
}

func encodePhoneNumber(phone string) string {
	if (len(phone) % 2) != 0 {
		phone += "F"

	}
	str := []rune(phone)
	for i := 0; i < len(str); i += 2 {
		str[i], str[i+1] = str[i+1], str[i]
	}
	return string(str)
}

func createUDH(slices int) []string {
	result := make([]string, slices)
	IED1 := fmt.Sprintf("%02X", rand.Intn(255))
	base := "05" + "00" + "03" + IED1 + fmt.Sprintf("%02X", slices)
	for i := 0; i < slices; i++ {
		result[i] = base + fmt.Sprintf("%02X", i)
	}
	return result
}

func getNextTpmr() string {
	tpmr++
	if tpmr == 256 {
		tpmr = 0
	}
	return fmt.Sprintf("%02X", tpmr)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println(createUDH(3))
}
