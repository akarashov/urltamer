package service

import (
	"github.com/akarashov/urltamer/internal/model"
)

// Observer defines the interface for audit event observers.
type Observer interface {
	Update(event model.AuditEvent)
	GetID() string
}

// Subject defines the interface for audit event subjects.
type Subject interface {
	Register(Observer)
	Deregister(Observer)
	Notify(event model.AuditEvent)
}

// AuditSubject implements the Subject interface for audit events.
type AuditSubject struct {
	observers map[string]Observer
}

// NewAuditSubject creates a new AuditSubject.
func NewAuditSubject() *AuditSubject {
	return &AuditSubject{
		observers: make(map[string]Observer),
	}
}

// Register adds an observer to the subject.
func (a *AuditSubject) Register(observer Observer) {
	a.observers[observer.GetID()] = observer
}

// Deregister removes an observer from the subject.
func (a *AuditSubject) Deregister(observer Observer) {
	delete(a.observers, observer.GetID())
}

// Notify notifies all registered observers of an audit event.
func (a *AuditSubject) Notify(event model.AuditEvent) {
	for _, observer := range a.observers {
		go observer.Update(event)
	}
}
