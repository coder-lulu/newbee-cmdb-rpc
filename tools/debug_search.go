package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/attribute"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valueindextext"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/valuetext"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接配置
	dsn := "user:password@tcp(localhost:3306)/cmdb?charset=utf8mb4&parseTime=True&loc=Local"
	if len(os.Args) > 1 {
		dsn = os.Args[1]
	}

	// 创建数据库连接
	client, err := ent.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 调试参数 - 可以通过命令行参数自定义
	attrID := uint64(2)
	searchValue := "追击"
	typeID := uint64(7)

	// 通用搜索测试值
	generalSearchValue := "admin2"

	if len(os.Args) > 2 {
		searchValue = os.Args[2]
	}
	if len(os.Args) > 3 {
		generalSearchValue = os.Args[3]
	}

	fmt.Printf("=== 调试搜索问题 ===\n")
	fmt.Printf("搜索条件: AttrID=%d, Value='%s', TypeID=%d\n", attrID, searchValue, typeID)
	fmt.Printf("通用搜索值: '%s'\n\n", generalSearchValue)

	// 1. 检查属性信息
	fmt.Println("1. 检查属性信息:")
	attr, err := client.Attribute.Query().Where(attribute.IDEQ(attrID)).First(ctx)
	if err != nil {
		fmt.Printf("   错误: 找不到属性ID %d: %v\n", attrID, err)
	} else {
		fmt.Printf("   属性ID: %d\n", attr.ID)
		fmt.Printf("   属性名称: %s\n", attr.Name)
		fmt.Printf("   属性别名: %s\n", attr.Alias)
		fmt.Printf("   值类型: %s\n", attr.ValueType)
		fmt.Printf("   是否选择: %t\n", attr.IsChoice)
	}
	fmt.Println()

	// 2. 检查CI信息
	fmt.Println("2. 检查CI信息:")
	ciList, err := client.Cis.Query().Where(cis.TypeIDEQ(typeID)).All(ctx)
	if err != nil {
		fmt.Printf("   错误: 查询CI失败: %v\n", err)
	} else {
		fmt.Printf("   找到%d个类型为%d的CI:\n", len(ciList), typeID)
		for _, ci := range ciList {
			fmt.Printf("   - CI ID: %d, 状态: %d, 创建时间: %v\n", ci.ID, ci.Status, ci.CreatedAt)
		}
	}
	fmt.Println()

	// 3. 检查ValueText表 - 属性搜索
	fmt.Printf("3. 检查ValueText表 - 属性搜索 ('%s'):\n", searchValue)
	valueTextList, err := client.ValueText.Query().Where(valuetext.AttrIDEQ(attrID)).All(ctx)
	if err != nil {
		fmt.Printf("   错误: 查询ValueText失败: %v\n", err)
	} else {
		fmt.Printf("   属性ID %d 在ValueText表中有%d条记录:\n", attrID, len(valueTextList))
		for _, vt := range valueTextList {
			fmt.Printf("   - CI ID: %d, 值: '%s'\n", vt.CiID, vt.Value)
			// 检查是否匹配搜索值
			if vt.Value == searchValue {
				fmt.Printf("     ✅ 完全匹配搜索值!\n")
			} else if len(vt.Value) > 0 && len(searchValue) > 0 {
				if vt.Value[0] == searchValue[0] {
					fmt.Printf("     ⚠️  首字符匹配，但值不完全相同\n")
				}
			}
		}
	}
	fmt.Println()

	// 4. 检查ValueIndexText表 - 属性搜索
	fmt.Printf("4. 检查ValueIndexText表 - 属性搜索 ('%s'):\n", searchValue)
	valueIndexTextList, err := client.ValueIndexText.Query().Where(valueindextext.AttrIDEQ(attrID)).All(ctx)
	if err != nil {
		fmt.Printf("   错误: 查询ValueIndexText失败: %v\n", err)
	} else {
		fmt.Printf("   属性ID %d 在ValueIndexText表中有%d条记录:\n", attrID, len(valueIndexTextList))
		for _, vit := range valueIndexTextList {
			fmt.Printf("   - CI ID: %d, 值: '%s'\n", vit.CiID, vit.Value)
			// 检查是否匹配搜索值
			if vit.Value == searchValue {
				fmt.Printf("     ✅ 完全匹配搜索值!\n")
			} else if len(vit.Value) > 0 && len(searchValue) > 0 {
				if vit.Value[0] == searchValue[0] {
					fmt.Printf("     ⚠️  首字符匹配，但值不完全相同\n")
				}
			}
		}
	}
	fmt.Println()

	// 5. 检查通用搜索 - ValueText表
	fmt.Printf("5. 检查通用搜索 - ValueText表 (包含 '%s'):\n", generalSearchValue)
	generalTextResults, err := client.ValueText.Query().Where(valuetext.ValueContains(generalSearchValue)).All(ctx)
	if err != nil {
		fmt.Printf("   错误: ValueText通用搜索失败: %v\n", err)
	} else {
		fmt.Printf("   ValueText表中包含'%s'的记录数: %d\n", generalSearchValue, len(generalTextResults))
		for i, vt := range generalTextResults {
			if i >= 10 { // 最多显示10条
				fmt.Printf("   ... 还有 %d 条记录\n", len(generalTextResults)-10)
				break
			}
			fmt.Printf("   - CI ID: %d, Attr ID: %d, 值: '%s'\n", vt.CiID, vt.AttrID, vt.Value)
		}
	}
	fmt.Println()

	// 6. 检查通用搜索 - ValueIndexText表
	fmt.Printf("6. 检查通用搜索 - ValueIndexText表 (包含 '%s'):\n", generalSearchValue)
	generalIndexTextResults, err := client.ValueIndexText.Query().Where(valueindextext.ValueContains(generalSearchValue)).All(ctx)
	if err != nil {
		fmt.Printf("   错误: ValueIndexText通用搜索失败: %v\n", err)
	} else {
		fmt.Printf("   ValueIndexText表中包含'%s'的记录数: %d\n", generalSearchValue, len(generalIndexTextResults))
		for i, vit := range generalIndexTextResults {
			if i >= 10 { // 最多显示10条
				fmt.Printf("   ... 还有 %d 条记录\n", len(generalIndexTextResults)-10)
				break
			}
			fmt.Printf("   - CI ID: %d, Attr ID: %d, 值: '%s'\n", vit.CiID, vit.AttrID, vit.Value)
		}
	}
	fmt.Println()

	// 7. 检查精确匹配
	fmt.Printf("7. 检查精确匹配 ('%s'):\n", searchValue)

	// ValueText精确匹配
	vtExact, err := client.ValueText.Query().
		Where(valuetext.AttrIDEQ(attrID), valuetext.ValueEQ(searchValue)).
		All(ctx)
	if err != nil {
		fmt.Printf("   ValueText精确匹配查询失败: %v\n", err)
	} else {
		fmt.Printf("   ValueText精确匹配结果: %d条\n", len(vtExact))
		for _, vt := range vtExact {
			fmt.Printf("   - CI ID: %d\n", vt.CiID)
		}
	}

	// ValueIndexText精确匹配
	vitExact, err := client.ValueIndexText.Query().
		Where(valueindextext.AttrIDEQ(attrID), valueindextext.ValueEQ(searchValue)).
		All(ctx)
	if err != nil {
		fmt.Printf("   ValueIndexText精确匹配查询失败: %v\n", err)
	} else {
		fmt.Printf("   ValueIndexText精确匹配结果: %d条\n", len(vitExact))
		for _, vit := range vitExact {
			fmt.Printf("   - CI ID: %d\n", vit.CiID)
		}
	}
	fmt.Println()

	// 8. 检查字符编码问题
	fmt.Println("8. 检查字符编码:")
	fmt.Printf("   搜索值 '%s' 的字节数组: %v\n", searchValue, []byte(searchValue))
	fmt.Printf("   搜索值长度: %d 字符, %d 字节\n", len([]rune(searchValue)), len(searchValue))
	fmt.Printf("   通用搜索值 '%s' 的字节数组: %v\n", generalSearchValue, []byte(generalSearchValue))
	fmt.Printf("   通用搜索值长度: %d 字符, %d 字节\n", len([]rune(generalSearchValue)), len(generalSearchValue))

	// 检查数据库中值的编码
	if len(valueTextList) > 0 {
		for i, vt := range valueTextList {
			if i >= 3 { // 只检查前3条
				break
			}
			fmt.Printf("   数据库值 '%s' 的字节数组: %v\n", vt.Value, []byte(vt.Value))
		}
	}

	if len(valueIndexTextList) > 0 {
		for i, vit := range valueIndexTextList {
			if i >= 3 { // 只检查前3条
				break
			}
			fmt.Printf("   索引表值 '%s' 的字节数组: %v\n", vit.Value, []byte(vit.Value))
		}
	}

	// 9. 总结报告
	fmt.Printf("\n=== 总结报告 ===\n")
	fmt.Printf("属性搜索 ('%s'):\n", searchValue)
	fmt.Printf("  - ValueText表匹配: %d条\n", len(vtExact))
	fmt.Printf("  - ValueIndexText表匹配: %d条\n", len(vitExact))

	fmt.Printf("通用搜索 ('%s'):\n", generalSearchValue)
	fmt.Printf("  - ValueText表包含: %d条\n", len(generalTextResults))
	fmt.Printf("  - ValueIndexText表包含: %d条\n", len(generalIndexTextResults))

	totalGeneral := len(generalTextResults) + len(generalIndexTextResults)
	if totalGeneral > 0 {
		fmt.Printf("  ✅ 通用搜索应该能找到数据 (总计: %d条)\n", totalGeneral)
	} else {
		fmt.Printf("  ❌ 通用搜索没有找到数据，可能需要检查数据存储位置\n")
	}

	fmt.Println("\n=== 调试完成 ===")
	fmt.Println("使用方法: go run tools/debug_search.go [数据库连接字符串] [属性搜索值] [通用搜索值]")
	fmt.Println("例如: go run tools/debug_search.go 'user:pass@tcp(localhost:3306)/cmdb?charset=utf8mb4&parseTime=True&loc=Local' '追击' 'admin2'")
}
