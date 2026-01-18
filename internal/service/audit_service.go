package service

import (
	"github.com/akarashov/urltamer/internal/model"
)

// определяем интерфейс наблюдателя для аудита
type Observer interface {
	Update(event model.AuditEvent)
	GetID() string
}

// определяем интерфейс обработчика событий для аудита
type Subject interface {
	Register(Observer)
	Deregister(Observer)
	Notify(event model.AuditEvent)
}

// реализация обработчика для событий аудита
type AuditSubject struct {
	observers map[string]Observer
}

func NewAuditSubject() *AuditSubject {
	return &AuditSubject{
		observers: make(map[string]Observer),
	}
}

func (a *AuditSubject) Register(observer Observer) {
	a.observers[observer.GetID()] = observer
}

func (a *AuditSubject) Deregister(observer Observer) {
	delete(a.observers, observer.GetID())
}

func (a *AuditSubject) Notify(event model.AuditEvent) {
	for _, observer := range a.observers {
		go observer.Update(event)
	}
}
