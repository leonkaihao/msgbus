package inproc

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"
)

type Integer struct {
	a int
}

func matchSubMapper(t *testing.T, sm *subMapper[Integer], topic string, expect int) {
	nr := sm.Check(strings.Split(topic, "."))
	if len(nr) != expect {
		t.Errorf("topic %v expects %v matches but got %v", topic, expect, len(nr))
	}
}

func TestSubMapper(t *testing.T) {
	sm := newSubMapper[Integer]()
	sm.Add([]string{">"}, &Integer{0})
	sm.Add([]string{"reading", ">"}, &Integer{1})
	sm.Add([]string{"event"}, &Integer{2})
	sm.Add([]string{"event", "*"}, &Integer{3})
	sm.Add([]string{"event", ">"}, &Integer{4})
	sm.Add([]string{"event", "zone"}, &Integer{5})
	sm.Add([]string{"event", "zone", "*"}, &Integer{6})
	sm.Add([]string{"*", "zone", "*"}, &Integer{7})
	sm.Add([]string{"event", "zone", ">"}, &Integer{8})
	sm.Add([]string{"event", "*", "entry"}, &Integer{9})
	sm.Add([]string{"event", "*", "*"}, &Integer{10})
	sm.Add([]string{"event", "*", ">"}, &Integer{11})
	sm.Add([]string{"event", "zone", "entry"}, &Integer{12})
	sm.Add([]string{"*", "*", "leave"}, &Integer{13})
	sm.Add([]string{"event", "*", "entry", "*"}, &Integer{14})
	sm.Add([]string{"event", "*", "entry", ">"}, &Integer{15})
	sm.Add([]string{"event", "zone", "entry", ">"}, &Integer{16})
	sm.Add([]string{"event", "zone", "entry", "cancel"}, &Integer{17})
	sm.Add([]string{"event", "zone", "entry", "cancel", "*"}, &Integer{18})
	sm.Add([]string{"event", "zone", "entry", "cancel", ">"}, &Integer{19})
	matchSubMapper(t, sm, "event", 2)                   // 0,2
	matchSubMapper(t, sm, "event.zone", 4)              // 0,3,4,5
	matchSubMapper(t, sm, "event.zone.entry", 9)        // 0,4,6,7,8,9,10,11,12
	matchSubMapper(t, sm, "event.zone.leave", 8)        // 0,4,6,7,8,10,11,13
	matchSubMapper(t, sm, "event.zone.entry.cancel", 8) // 0,4,8,11,14,15,16,17
	sm.Remove([]string{"event", "zone", "entry", "cancel", ">"})
	sm.Remove([]string{"event", "zone", "entry", "cancel"})
	sm.Remove([]string{">"})
	matchSubMapper(t, sm, "event.zone.entry.cancel", 6) // 4,8,11,14,15,16
}

type Consumers []model.Consumer

func (c Consumers) Len() int           { return len(c) }
func (c Consumers) Less(i, j int) bool { return c[i].Name() < c[j].Name() }
func (c Consumers) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func matchConsumer(t *testing.T, tm *TopicMapper, topic string, expect []model.Consumer) {
	cmrs, err := tm.matchConsumers(topic)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(cmrs) != len(expect) {
		t.Fatalf("match topic %v, expect %v match , but got %v match", topic, len(expect), len(cmrs))
		return
	}
	sort.Sort(Consumers(cmrs))
	sort.Sort(Consumers(expect))
	for i := 0; i < len(cmrs); i++ {
		if cmrs[i] != expect[i] {
			t.Fatalf("expect match (%v, %v, %v), but got (%v, %v, %v)", expect[i].ID(), expect[i].Sub(), expect[i].Group(), cmrs[i].ID(), cmrs[i].Sub(), cmrs[i].Group())
			return
		}
	}
}

func TestTopicMapperDifferentSubGroup(t *testing.T) {
	// differnt sub and group should find respective consumers
	tm := NewTopicMapper()
	cm1 := common.NewConsumerBase("cm1", "sub1", "group1")
	cm2 := common.NewConsumerBase("cm2", "sub2", "group2")
	tm.Subscribe(cm1, make(chan<- model.Messager))
	tm.Subscribe(cm2, make(chan<- model.Messager))
	matchConsumer(t, tm, "sub1", []model.Consumer{cm1})
	matchConsumer(t, tm, "sub2", []model.Consumer{cm2})
	matchConsumer(t, tm, "sub3", []model.Consumer{})
}

func TestTopicMapperSameSubGroup(t *testing.T) {
	// same sub and group should find a consumer alternately in a group
	tm := NewTopicMapper()
	cm1 := common.NewConsumerBase("cm1", "sub1", "group1")
	cm2 := common.NewConsumerBase("cm2", "sub1", "group1")
	tm.Subscribe(cm1, make(chan<- model.Messager))
	tm.Subscribe(cm2, make(chan<- model.Messager))
	matchConsumer(t, tm, "sub1", []model.Consumer{cm1})
	matchConsumer(t, tm, "sub1", []model.Consumer{cm2})
	matchConsumer(t, tm, "sub1", []model.Consumer{cm1})
}

func TestTopicMapperSameSubDifferentGroup(t *testing.T) {
	// same sub and different group should find consumers in different group alternately
	tm := NewTopicMapper()
	cm11 := common.NewConsumerBase("cm11", "sub1", "group1")
	cm21 := common.NewConsumerBase("cm21", "sub1", "group2")
	cm22 := common.NewConsumerBase("cm22", "sub1", "group2")
	cm31 := common.NewConsumerBase("cm31", "sub1", "group3")
	cm32 := common.NewConsumerBase("cm32", "sub1", "group3")
	cm33 := common.NewConsumerBase("cm33", "sub1", "group3")
	tm.Subscribe(cm11, make(chan<- model.Messager))
	tm.Subscribe(cm21, make(chan<- model.Messager))
	tm.Subscribe(cm22, make(chan<- model.Messager))
	tm.Subscribe(cm31, make(chan<- model.Messager))
	tm.Subscribe(cm32, make(chan<- model.Messager))
	tm.Subscribe(cm33, make(chan<- model.Messager))
	matchConsumer(t, tm, "sub1", []model.Consumer{cm11, cm21, cm31})
	matchConsumer(t, tm, "sub1", []model.Consumer{cm11, cm22, cm32})
	matchConsumer(t, tm, "sub1", []model.Consumer{cm11, cm21, cm33})
	matchConsumer(t, tm, "sub1", []model.Consumer{cm11, cm22, cm31})
}

func TestTopicMapperUnsubSameGroup(t *testing.T) {
	//
	tm := NewTopicMapper()
	cm1 := common.NewConsumerBase("cm1", "sub1", "group1")
	cm2 := common.NewConsumerBase("cm2", "sub1", "group1")
	cm3 := common.NewConsumerBase("cm3", "sub1", "group1")
	cm4 := common.NewConsumerBase("cm4", "sub1", "group1")
	tm.Subscribe(cm1, make(chan<- model.Messager))
	tm.Subscribe(cm2, make(chan<- model.Messager))
	tm.Subscribe(cm3, make(chan<- model.Messager))
	tm.Subscribe(cm4, make(chan<- model.Messager))
	matchConsumer(t, tm, "sub1", []model.Consumer{cm1})
	if err := tm.UnSubscribe(cm2); err != nil {
		t.Fatal(err)
	}
	matchConsumer(t, tm, "sub1", []model.Consumer{cm3})
	if err := tm.UnSubscribe(cm4); err != nil {
		t.Fatal(err)
	}
	matchConsumer(t, tm, "sub1", []model.Consumer{cm1})
}

func TestTopicMapperMixedSubInGroup(t *testing.T) {
	tm := NewTopicMapper()
	cm11 := common.NewConsumerBase("cm11", "sub1", "group1")
	cm12 := common.NewConsumerBase("cm12", "sub1", "group1")
	cm13 := common.NewConsumerBase("cm13", "sub1", "group2")
	cm14 := common.NewConsumerBase("cm14", "sub1", "group3")

	cm21 := common.NewConsumerBase("cm21", "sub2", "group2")
	cm22 := common.NewConsumerBase("cm22", "sub2", "group2")
	cm23 := common.NewConsumerBase("cm23", "sub2", "group1")
	cm24 := common.NewConsumerBase("cm24", "sub2", "group3")

	cm31 := common.NewConsumerBase("cm31", "sub3", "group3")
	cm32 := common.NewConsumerBase("cm32", "sub3", "group3")
	cm33 := common.NewConsumerBase("cm33", "sub3", "group1")
	cm34 := common.NewConsumerBase("cm34", "sub3", "group2")
	tm.Subscribe(cm11, make(chan<- model.Messager))
	tm.Subscribe(cm12, make(chan<- model.Messager))
	tm.Subscribe(cm13, make(chan<- model.Messager))
	tm.Subscribe(cm14, make(chan<- model.Messager))
	tm.Subscribe(cm21, make(chan<- model.Messager))
	tm.Subscribe(cm22, make(chan<- model.Messager))
	tm.Subscribe(cm23, make(chan<- model.Messager))
	tm.Subscribe(cm24, make(chan<- model.Messager))
	tm.Subscribe(cm31, make(chan<- model.Messager))
	tm.Subscribe(cm32, make(chan<- model.Messager))
	tm.Subscribe(cm33, make(chan<- model.Messager))
	tm.Subscribe(cm34, make(chan<- model.Messager))
	matchConsumer(t, tm, "sub1", []model.Consumer{cm11, cm13, cm14})
	matchConsumer(t, tm, "sub1", []model.Consumer{cm12, cm13, cm14})
	matchConsumer(t, tm, "sub2", []model.Consumer{cm21, cm23, cm24})
	matchConsumer(t, tm, "sub2", []model.Consumer{cm22, cm23, cm24})
	matchConsumer(t, tm, "sub3", []model.Consumer{cm31, cm33, cm34})
	matchConsumer(t, tm, "sub3", []model.Consumer{cm32, cm33, cm34})
	tm.UnSubscribe(cm11)
	tm.UnSubscribe(cm22)
	tm.UnSubscribe(cm33)
	matchConsumer(t, tm, "sub1", []model.Consumer{cm12, cm13, cm14})
	matchConsumer(t, tm, "sub2", []model.Consumer{cm21, cm23, cm24})
	matchConsumer(t, tm, "sub3", []model.Consumer{cm31, cm34})
}

func BenchmarkMatchGroup10(b *testing.B) {
	tm := NewTopicMapper()
	for i := 0; i < 10; i++ {
		cm := common.NewConsumerBase(fmt.Sprintf("cm%3v", i), "sub1", fmt.Sprintf("group%3v", i))
		tm.Subscribe(cm, make(chan<- model.Messager))
	}
	for i := 0; i < 1000000; i++ {
		tm.matchConsumers("sub1")
	}
}

func BenchmarkMatchGroup100(b *testing.B) {
	tm := NewTopicMapper()
	for i := 0; i < 100; i++ {
		cm := common.NewConsumerBase(fmt.Sprintf("cm%3v", i), "sub1", fmt.Sprintf("group%3v", i))
		tm.Subscribe(cm, make(chan<- model.Messager))
	}
	for i := 0; i < 1000000; i++ {
		tm.matchConsumers("sub1")
	}
}

func BenchmarkMatchGroup1000(b *testing.B) {
	tm := NewTopicMapper()
	for i := 0; i < 1000; i++ {
		cm := common.NewConsumerBase(fmt.Sprintf("cm%3v", i), "sub1", fmt.Sprintf("group%3v", i))
		tm.Subscribe(cm, make(chan<- model.Messager))
	}
	for i := 0; i < 1000000; i++ {
		tm.matchConsumers("sub1")
	}
}

func BenchmarkMatchTopic1000(b *testing.B) {
	tm := NewTopicMapper()
	for i := 0; i < 100; i++ {
		cm := common.NewConsumerBase(fmt.Sprintf("cm%3v", i), fmt.Sprintf("sub%3v", i), fmt.Sprintf("group%3v", i))
		tm.Subscribe(cm, make(chan<- model.Messager))
	}
	for i := 0; i < 1000000; i++ {
		tm.matchConsumers("sub099")
	}
}
