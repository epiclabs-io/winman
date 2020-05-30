package winman

type Stack []interface{}

func (s *Stack) Push(newItem interface{}) {
	if newItem == nil {
		panic("cannot add nil item to stack")
	}
	for _, item := range *s {
		if item == newItem {
			return
		}
	}
	*s = append(*s, newItem)
}

func (s *Stack) Pop() interface{} {
	lenItems := len(*s)
	if lenItems == 0 {
		return nil
	}
	var item interface{}
	item, *s = (*s)[lenItems-1], (*s)[:lenItems-1]
	return item
}

func (s *Stack) Remove(item interface{}) {
	i := s.IndexOf(item)
	if i != -1 {
		*s = append((*s)[:i], (*s)[i+1:]...)
	}
}

func (s Stack) Item(i int) interface{} {
	if i < 0 || i >= len(s) {
		return nil
	}
	return s[i]
}

func (s Stack) IndexOf(searchItem interface{}) int {
	for i, item := range s {
		if item == searchItem {
			return i
		}
	}
	return -1
}

func (s *Stack) Move(item interface{}, targetIndex int) {
	oldIndex := s.IndexOf(item)
	if oldIndex == -1 {
		return
	}
	lenS := len(*s)

	if targetIndex < 0 || targetIndex >= lenS {
		targetIndex = lenS - 1
	}

	newStack := make([]interface{}, lenS)
	for i, j := 0, 0; i < lenS; j++ {
		if j == oldIndex {
			j++
		}
		if i == targetIndex {
			j--
		} else {
			newStack[i] = (*s)[j]
		}
		i++
	}

	newStack[targetIndex] = item
	*s = newStack
}

func (s Stack) Find(f func(item interface{}) bool) interface{} {
	for i := len(s) - 1; i >= 0; i-- {
		item := s[i]
		if f(item) {
			return item
		}
	}
	return nil
}
