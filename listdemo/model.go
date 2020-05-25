package main

import (
	"fmt"
	"os"
)

type rowModel struct {
	name string
	uuid string
}

type dataModel struct {
	rows []rowModel
}

func (m *dataModel) Len() int { return len(m.rows) }
func (m *dataModel) Equal(a interface{}, b interface{}) bool {
	return a.(rowModel) == b.(rowModel)
}
func (m *dataModel) GetItem(position int) (interface{}, error) {
	if position >= len(m.rows) {
		return nil, os.ErrNotExist
	}
	return m.rows[position], nil
}

func (m *dataModel) delete(uuid string) {
	delIdx := -1
	for i, r := range m.rows {
		if r.uuid == uuid {
			delIdx = i
			break
		}
	}

	if delIdx >= 0 {
		m.rows = append(m.rows[:delIdx], m.rows[delIdx+1:]...)
	}
}

var lastNum int

func (m *dataModel) add() {
	m.rows = append(m.rows, rowModel{
		uuid: fmt.Sprintf("row-new-%d", lastNum),
	})
	lastNum++
}

func (m *dataModel) updateName(uuid, name string) {
	fmt.Printf("UpdateName(%q, %q)\n", uuid, name)
	for i, r := range m.rows {
		if r.uuid == uuid {
			m.rows[i].name = name
		}
	}
}
