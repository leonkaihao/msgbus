package inproc

import (
	"sort"
	"testing"

	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"
)

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
