package regolt

type EventController[T any] struct {
	ls map[int]func(T)
	id int
}

func (ec *EventController[T]) inc() int {
	ec.id++
	return ec.id
}

type Subscription[T any] struct {
	Controller *EventController[T]
	ID         int
}

func (s *Subscription[T]) Delete() {
	delete(s.Controller.ls, s.ID)
}

func (ec *EventController[T]) Listen(f func(T)) *Subscription[T] {
	i := ec.inc()
	ec.ls[i] = f
	return &Subscription[T]{Controller: ec, ID: i}
}

func (ec *EventController[T]) Override(f func(T)) *Subscription[T] {
	clear(ec.ls)
	return ec.Listen(f)
}

func (ec *EventController[T]) Emit(t T) *EventController[T] {
	for _, g := range ec.ls {
		g(t)
	}
	return ec
}

func (ec *EventController[T]) EmitInGoroutines(t T) *EventController[T] {
	if len(ec.ls) == 0 {
		return ec
	}
	go ec.Emit(t)
	return ec
}

func (ec *EventController[T]) EmitAndCall(t T, f func(T)) *EventController[T] {
	if len(ec.ls) == 0 {
		f(t)
		return ec
	}
	go func(x *EventController[T], y T, z func(T)) {
		x.Emit(y)
		z(y)
	}(ec, t, f)
	return ec
}

func NewEventController[T any]() *EventController[T] {
	return &EventController[T]{
		ls: map[int]func(T){},
	}
}
