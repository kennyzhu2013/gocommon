/*
* Copyright(C),2019-2029, email: 277251257@qq.com
* Author:  kennyzhu
* Version: 1.0.0
* Date:    2021/3/19 16:27
* Description:
*
 */
package main

import (
	"common/util/process"
	"errors"
	"fmt"
)

func main() {
	var p1 = process.Resolve(123)
	var p2 = process.Resolve("Hello, World")
	var p3 = process.Resolve([]string{"one", "two", "three"})

	results, _ := process.All(p1, p2, p3).Await()
	fmt.Println(results)

	var p = process.New(func(resolve func(process.Any), reject func(error)) {
		// Do something asynchronously.
		const sum = 2 + 2

		// If your work was successful call resolve() passing the result.
		if sum == 4 {
			resolve(sum)
			return
		}

		// If you encountered an error call reject() passing the error.
		if sum != 4 {
			reject(errors.New("2 + 2 doesnt't equal 4"))
			return
		}
	}).
		Then(func(data process.Any) process.Any {
			fmt.Println("The result is:", data)
			return data.(int) + 1
		}).
		Then(func(data process.Any) process.Any {
			fmt.Println("The new result is:", data)
			return nil
		}).
		Catch(func(err error) error {
			fmt.Println("Error during execution:", err.Error())
			return nil
		})

	// Since handlers are executed asynchronously you can wait for them.
	p.Await()
}
