package poker24

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"
	"testing"
)

type ONE struct {
	com     []int
	formula string
}

func Test24(t *testing.T) {
	res := []ONE{}
	mm := map[string]ONE{}
	for i := 1; i <= 10; i++ {
		for j := 1; j <= 10; j++ {
			for m := 1; m <= 10; m++ {
				for n := 1; n <= 10; n++ {
					s := []int{i, j, m, n}
					slices.Sort(s)
					if judgePoint24(s) {
						encode := strconv.Itoa(s[0]) + strconv.Itoa(s[1]) + strconv.Itoa(s[2]) + strconv.Itoa(s[3])
						if _, ok := mm[encode]; !ok {
							one := ONE{
								com:     s,
								formula: resString[len(resString)-1],
							}
							res = append(res, one)
							mm[encode] = one
						}
					}
				}
			}
		}
	}
	fmt.Println("一共 ", len(res), " 个")
	for _, one := range res {
		fmt.Println(one.com)
		fmt.Println(one.formula)
	}
	sqlPattern := "with cte_flatten_cxs as (select id, group_concat(c order by c SEPARATOR \"\") as flatten_c1c2c3c4 from ((select id, c1 as c from `poker24`.`cards`) union all (select id, c2 as c from `poker24`.`cards`) union all(select id, c3 as c from `poker24`.`cards`) union all (select id,c4 as c from `poker24`.`cards`)) allrows group by id) select %s from cte_flatten_cxs join `poker24`.`cards` using (id);"
	casewhen := genOneCaseWhenfunc(res, 0)
	select_fields := "id, c1, c2, c3, c4, (" + casewhen + ") as result"
	query := fmt.Sprintf(sqlPattern, select_fields)
	fmt.Println(query)

	// insert test data:
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/poker24")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	fmt.Println("Success!")
	db.Exec("delete from `poker24`.`cards`")
	for _, one := range res {
		_, err := db.Exec(fmt.Sprintf("insert into `poker24`.`cards` values (default, %d, %d, %d, %d)", one.com[0], one.com[1], one.com[2], one.com[3]))
		if err != nil {
			panic(err.Error())
		}
	}
}

var oneCaseWhenTemplate = "case when (flatten_c1c2c3c4 = %s) then \"%s\" else (%s) end"

func genOneCaseWhenfunc(ones []ONE, idx int) string {
	s := ones[idx].com
	result := ones[idx].formula
	encode := strconv.Itoa(s[0]) + strconv.Itoa(s[1]) + strconv.Itoa(s[2]) + strconv.Itoa(s[3])
	if idx == len(ones)-1 {
		return fmt.Sprintf(oneCaseWhenTemplate, encode, result, "null")
	}
	elseString := genOneCaseWhenfunc(ones, idx+1)
	return fmt.Sprintf(oneCaseWhenTemplate, encode, result, elseString)
}

const (
	TARGET      = 24
	EPSILON     = 1e-6
	NONE     OP = -1
	ADD      OP = 0
	MULTIPLY OP = 1
	SUBTRACT OP = 2
	DIVIDE   OP = 3
)

type OP int

func (alreadyOP OP) priorTo(currentOP OP, alreadyOPIsLeft bool) bool {
	// 1 ? (xxx)
	if alreadyOP == NONE {
		return true
	}
	// (xxx) ? 1
	if currentOP == NONE {
		return false
	}
	// both are divide, like 10/(2/6)-6
	if alreadyOP == DIVIDE && currentOP == DIVIDE {
		if alreadyOPIsLeft {
			return true
		} else {
			return false
		}
	}
	// if two sub together, like (8-6-2)*6
	if alreadyOP == SUBTRACT && currentOP == SUBTRACT {
		if alreadyOPIsLeft {
			return true
		} else {
			return false
		}
	}
	// if first is sub and second is add like: 9*4-8+4, the latter should be enclosed as 9*4-(8+4)
	if alreadyOP == ADD && currentOP == SUBTRACT {
		if alreadyOPIsLeft {
			return true
		} else {
			return false
		}
	}
	if alreadyOP == MULTIPLY || alreadyOP == DIVIDE {
		return true
	}
	if currentOP == MULTIPLY || currentOP == DIVIDE {
		return false
	}
	return true
}

func (op OP) string() string {
	switch op {
	case ADD:
		return "+"
	case MULTIPLY:
		return "*"
	case SUBTRACT:
		return "-"
	case DIVIDE:
		return "/"
	}
	return ""
}

func fromString(f1, f2 TmpFloat, currentOp OP) string {
	var res string
	//(8-6-2)*6
	if !f1.lastOP.priorTo(currentOp, true) {
		res += "("
		res += f1.from
		res += ")"
	} else {
		res += f1.from
	}
	res += currentOp.string()
	if !f2.lastOP.priorTo(currentOp, false) {
		res += "("
		res += f2.from
		res += ")"
	} else {
		res += f2.from
	}
	return res
}

func judgePoint24(nums []int) bool {
	slices.Sort(nums)
	list := []TmpFloat{}
	for _, num := range nums {
		list = append(list, TmpFloat{value: float64(num), from: strconv.Itoa(num), lastOP: NONE})
	}
	return solve(list)
}

type TmpFloat struct {
	value  float64
	lastOP OP
	from   string
}

var resString []string

func solve(list []TmpFloat) bool {
	if len(list) == 0 {
		return false
	}
	if len(list) == 1 {
		if abs(list[0].value-TARGET) < EPSILON {
			resString = append(resString, list[0].from)
			return true
		}
	}
	size := len(list)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if i != j {
				list2 := []TmpFloat{}
				for k := 0; k < size; k++ {
					if k != i && k != j {
						list2 = append(list2, list[k])
					}
				}
				for k := OP(0); k < 4; k++ {
					if k < 2 && i < j {
						continue
					}
					switch k {
					case ADD:
						list2 = append(list2, TmpFloat{
							value:  list[i].value + list[j].value,
							from:   fromString(list[i], list[j], ADD),
							lastOP: ADD,
						})
					case MULTIPLY:
						list2 = append(list2, TmpFloat{
							value:  list[i].value * list[j].value,
							from:   fromString(list[i], list[j], MULTIPLY),
							lastOP: MULTIPLY,
						})
					case SUBTRACT:
						if list[i].value-list[j].value < 0 {
							list2 = append(list2, TmpFloat{
								value:  list[j].value - list[i].value,
								from:   fromString(list[j], list[i], SUBTRACT),
								lastOP: SUBTRACT,
							})
						} else {
							list2 = append(list2, TmpFloat{
								value:  list[i].value - list[j].value,
								from:   fromString(list[i], list[j], SUBTRACT),
								lastOP: SUBTRACT,
							})
						}
					case DIVIDE:
						if abs(list[j].value) < EPSILON {
							continue
						} else {
							list2 = append(list2, TmpFloat{
								value:  list[i].value / list[j].value,
								from:   fromString(list[i], list[j], DIVIDE),
								lastOP: DIVIDE,
							})
						}
					}
					if solve(list2) {
						return true
					}
					list2 = list2[:len(list2)-1]
				}
			}
		}
	}
	return false
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
