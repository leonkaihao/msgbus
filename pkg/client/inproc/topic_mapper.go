package inproc

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/leonkaihao/msgbus/pkg/model"
)

type subConsumers struct {
	Consumers []model.Consumer
	CurIndex  int
}

type consumerGroup struct {
	name    string
	mapping map[string]*subConsumers // sub: consumers
}

type subMapper[T any] struct {
	sub map[string]*subMapper[T]
	val *T
}

func newSubMapper[T any]() *subMapper[T] {
	return &subMapper[T]{
		sub: make(map[string]*subMapper[T]),
	}
}

func (sm *subMapper[T]) Add(tail []string, val *T) {
	if len(tail) == 0 {
		sm.val = val
		return
	}
	if mp, ok := sm.sub[tail[0]]; ok {
		mp.Add(tail[1:], val)
	} else {
		mp = newSubMapper[T]()
		mp.Add(tail[1:], val)
		sm.sub[tail[0]] = mp
	}
}

func (sm *subMapper[T]) TraverseCreate(tail []string) *subMapper[T] {
	if len(tail) == 0 {
		return sm
	}
	if mp, ok := sm.sub[tail[0]]; ok {
		return mp.TraverseCreate(tail[1:])
	} else {
		mp = newSubMapper[T]()
		smNew := mp.TraverseCreate(tail[1:])
		sm.sub[tail[0]] = mp
		return smNew
	}
}

func (sm *subMapper[T]) Remove(tail []string) {
	if len(tail) == 1 {
		delete(sm.sub, tail[0])
		return
	}
	if mp, ok := sm.sub[tail[0]]; ok {
		mp.Remove(tail[1:])
	}
}

func (sm *subMapper[T]) Check(tail []string) []*T {
	var (
		result = []*T{}
	)
	if len(tail) == 0 {
		if sm.val != nil {
			result = []*T{sm.val}
		}
	} else {
		if s, ok := sm.sub[tail[0]]; ok {
			ss := s.Check(tail[1:])
			result = append(ss, result...)
		}
		if s, ok := sm.sub["*"]; ok {
			ss := s.Check(tail[1:])
			result = append(ss, result...)
		}
		if s, ok := sm.sub[">"]; ok {
			result = append(result, s.val)
		}
	}
	return result
}

type TopicMapper struct {
	sync.RWMutex
	consumerGroups  map[string]*consumerGroup // group_name: object
	subMapping      map[string]map[string]int // sub: [group_name:count, ...]
	consumerChanMap map[model.Consumer]chan<- model.Messager
}

func NewTopicMapper() *TopicMapper {
	return &TopicMapper{
		consumerGroups:  make(map[string]*consumerGroup),
		subMapping:      make(map[string]map[string]int),
		consumerChanMap: make(map[model.Consumer]chan<- model.Messager),
	}
}

func (tm *TopicMapper) Subscribe(csmr model.Consumer, chn chan<- model.Messager) {
	tm.Lock()
	defer tm.Unlock()
	group, ok := tm.consumerGroups[csmr.Group()]
	if !ok {
		group = &consumerGroup{
			name:    csmr.Group(),
			mapping: make(map[string]*subConsumers),
		}
		tm.consumerGroups[csmr.Group()] = group
	}
	subm, ok := tm.subMapping[csmr.Sub()]
	if !ok {
		subm = map[string]int{csmr.Group(): 1}
		tm.subMapping[csmr.Sub()] = subm
	} else {
		subm[csmr.Group()]++
	}
	subcsmrs, ok := group.mapping[csmr.Sub()]
	if !ok {
		subcsmrs = &subConsumers{
			Consumers: []model.Consumer{},
		}
		group.mapping[csmr.Sub()] = subcsmrs
	}
	subcsmrs.Consumers = append(subcsmrs.Consumers, csmr)
	tm.consumerChanMap[csmr] = chn
}

func (tm *TopicMapper) UnSubscribe(csmr model.Consumer) error {
	tm.Lock()
	defer tm.Unlock()
	group, ok := tm.consumerGroups[csmr.Group()]
	if !ok {
		return fmt.Errorf("failed to unsubscribe consumer (%v,%v,%v), group not registered", csmr.ID(), csmr.Sub(), csmr.Group())
	}
	subcsmrs, ok := group.mapping[csmr.Sub()]
	if !ok {
		return fmt.Errorf("failed to unsubscribe consumer (%v,%v,%v), sub topic not registered", csmr.ID(), csmr.Sub(), csmr.Group())
	}
	for i, c := range subcsmrs.Consumers {
		if csmr == c {
			subcsmrs.Consumers = append(subcsmrs.Consumers[:i], subcsmrs.Consumers[i+1:]...)
			break
		}
	}
	if len(subcsmrs.Consumers) == 0 {
		delete(group.mapping, csmr.Sub())
	} else {
		subcsmrs.CurIndex = subcsmrs.CurIndex % len(subcsmrs.Consumers)
	}
	if len(group.mapping) == 0 {
		delete(tm.consumerGroups, csmr.Group())
	}

	groups, ok := tm.subMapping[csmr.Sub()]
	if ok {
		if count, ok := groups[csmr.Group()]; ok {
			count--
			groups[csmr.Group()] = count
			if count == 0 {
				delete(groups, csmr.Group())
			}
		}
	}
	return nil
}

func (tm *TopicMapper) Dispatch(msg model.Messager) error {
	tm.RLock()
	defer tm.RUnlock()
	csmrs, err := tm.matchConsumers(msg.Topic())
	if err != nil {
		return err
	}
	for _, csmr := range csmrs {
		ch, ok := tm.consumerChanMap[csmr]
		if !ok {
			return fmt.Errorf("consumer %v is not subscribed", csmr.ID())
		}
		ch <- msg
	}
	return nil
}

func (tm *TopicMapper) matchConsumers(topic string) ([]model.Consumer, error) {
	csmrs := []model.Consumer{}
	for sub, groups := range tm.subMapping {
		subExp := regexp.MustCompile(sub)
		if subExp.MatchString(topic) {
			for groupName := range groups {
				cgroup, ok := tm.consumerGroups[groupName]
				if !ok {
					return nil, fmt.Errorf("failed to get consumer group %v", groupName)
				}
				cs, ok := cgroup.mapping[sub]
				if !ok {
					return nil, fmt.Errorf("failed to get topic %v from consumer group %v", sub, groupName)
				}
				c := cs.Consumers[cs.CurIndex]
				cs.CurIndex = (cs.CurIndex + 1) % len(cs.Consumers)
				csmrs = append(csmrs, c)
			}
		}
	}
	return csmrs, nil
}
