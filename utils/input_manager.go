
package utils

import (
	"bufio"
	"fmt"
	

	"strings"
	"sync"
)

// InputManager gerencia toda entrada do usuário
type InputManager struct {
	Scanner *bufio.Scanner
	Mutex   sync.Mutex
	Active  bool
}


// ReadString lê uma string do stdin de forma thread-safe
func (im *InputManager) ReadString() string {
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	
	if !im.Active {
		return ""
	}
	
	if im.Scanner.Scan() {
		return strings.TrimSpace(im.Scanner.Text())
	}
	return ""
}

// ReadInt lê um inteiro do stdin de forma thread-safe
func (im *InputManager) ReadInt() (int, error) {
    if !im.Active {
        return 0, fmt.Errorf("input manager not active")
    }
    
    var value int
    _, err := fmt.Scanln(&value)
    return value, err
}

// WaitForEnter aguarda o usuário pressionar Enter
func (im *InputManager) WaitForEnter() {
	im.ReadString()
}

// Close desativa o input manager
func (im *InputManager) Close() {
	im.Mutex.Lock()
	defer im.Mutex.Unlock()
	im.Active = false
}