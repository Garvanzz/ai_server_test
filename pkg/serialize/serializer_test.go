package serialize

import (
	"fmt"
	"testing"
	"xfx/pkg/serialize/json"
)

type size struct {
	Chest int
	Waist int
	Hips  int
}

type beauty struct {
	Name   string
	Age    int
	Height float32
	Weight float32
	Size   *size
}

func (v beauty) Print() {
	fmt.Printf("Name = %v\n", v.Name)
	fmt.Printf("Age = %v\n", v.Age)
	fmt.Printf("Height = %v\n", v.Height)
	fmt.Printf("Weight = %v\n", v.Weight)
	fmt.Printf("Size.Chest = %v\n", v.Size.Chest)
	fmt.Printf("Size.Waist = %v\n", v.Size.Waist)
	fmt.Printf("Size.Hips = %v\n", v.Size.Hips)
}

func TestJson(t *testing.T) {
	v := beauty{
		Name:   "lisa",
		Age:    30,
		Height: 1.65,
		Weight: 55,
		Size: &size{
			Chest: 45,
			Waist: 30,
			Hips:  45,
		},
	}
	var js jsonSerializer = json.NewSerializer()
	data, _ := js.Marshal(&v)
	fmt.Printf("lisa: %s\n", string(data))

	var v2 beauty
	js.Unmarshal(data, &v2)

	// fmt.Printf("lisa 2: %v\n", v2)
	v2.Print()

	data2, _ := js.Marshal(&v2)
	fmt.Printf("lisa 2: %s\n", string(data2))
}
