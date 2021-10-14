/*
* Copyright(C),2019-2029, email: 277251257@qq.com
* Author:  kennyzhu
* Version: 1.0.0
* Date:    2021/3/19 16:32
* Description:
*
 */
package main

import "fmt"

func main() {
	maptest := make(map[int]int, 1000)
	for j := 1; j < 11; j++ {
		i := j * 100
		for ; i < j*100+100; i++ {
			maptest[i] = i
		}

		i = j * 10000
		for ; i < j*10000+10000; i++ {
			maptest[i] = i
		}
		fmt.Printf("[%d]:a, maptest size is : %d\n", j, len(maptest))

		i = j * 10000
		for ; i < j*10000+10000; i++ {
			delete(maptest, i)
		}
		fmt.Printf("[%d]:b, maptest size is : %d\n", j, len(maptest))
	}

	for j := 1; j < 11; j++ {
		i := j * 100
		for ; i < j*100+100; i++ {
			delete(maptest, i)
		}
	}
	fmt.Printf("[Last]:b, maptest size is : %d\n", len(maptest))

}
