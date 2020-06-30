package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "strings"
    "strconv"
    "sort"
    "flag"
)

func main() {
    
    // Переменные-параметры
    var fileName string
    //var fileName1 string
    var epochFrom int64
    var epochTo int64
    var slaTime int64
    var timestampCsvFieldNumber int
    var successCsvFieldNumber int
    var elapsedCsvFieldNumber int
    var showAllPercentiles bool
    
    // Обработка параметров командной строки
    fileNamePtr := flag.String("file", "", "Path to a jmeter csv log file, example: \"/home/username/volumes/jmeter/logs/logs.csv\"")
    epochFromPtr := flag.Int64("from", 0, "an int64 epoch timestamp, example: 1591365140000")
    epochToPtr := flag.Int64("to", 0, "an int64 epoch timestamp, example: 1591368740547")
    slaTimePtr := flag.Int64("sla", 0, "SLA in milliseconds, example: 200")
    timestampCsvFieldNumberPtr := flag.Int("epoch-num", 0, "a number of 'timestamp' field in CSV log file (numbering from zero), by default: 0")
    elapsedCsvFieldNumberPtr := flag.Int("elap-num", 1, "a number of 'elasped' field in CSV log file (numbering from zero), by default: 1")
    successCsvFieldNumberPtr := flag.Int("succ-num", 7, "a number of 'success' field in CSV log file (numbering from zero), by default: 7")
    showAllPercentilesPtr := flag.Bool("all", false, "use this flag to show time for each percentile from 1-st to 100-th")
    
    flag.Parse()
    fileName = *fileNamePtr //"/home/iv/doc/projects/presentation-90pct/logs.csv"
    epochFrom = *epochFromPtr
    epochTo = *epochToPtr
    timestampCsvFieldNumber = *timestampCsvFieldNumberPtr
    elapsedCsvFieldNumber = *elapsedCsvFieldNumberPtr
    successCsvFieldNumber = *successCsvFieldNumberPtr
    showAllPercentiles = *showAllPercentilesPtr
    slaTime = *slaTimePtr
    
    if( fileName == "" || epochFrom == 0 || epochTo == 0 || slaTime == 0  ) {
        fmt.Println("flags: -file, -from, -to, -sla - are requared!")
        fmt.Println("Example:")
        fmt.Println("./prc-counter -file /home/username/volumes/jmeter/logs/logs.csv -from 1591365140000 -to 1591368740547 -sla 200 -all")
        fmt.Println("use flag -h or -help to see more information")
        os.Exit(1)
    }
    
    // Открываем файл для чтения
    file, err := os.Open(fileName)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)

    // в этом цикле идет построчная обработка содержимого файла
    var totalRequests int64
    var slaAllowedRequests int64
    var prc[]int
    
    for scanner.Scan() {
        var line = scanner.Text()
        var splits = strings.Split(line, ",")
        
        var successStr = splits[successCsvFieldNumber]
        
        var epochStr = splits[timestampCsvFieldNumber]
        epoch, err := strconv.ParseInt(epochStr, 10, 64)
        if err != nil {
            fmt.Printf("Не смог распарсить строку:\n%s\nЗначение для перевода в число:%s\nРезультат парсинга:%d\n", line, epochStr, epoch)
        } else {
            var elapsedStr = splits[elapsedCsvFieldNumber]
            elapsed, _ := strconv.ParseInt(elapsedStr, 10, 32)
            if (epoch >= epochFrom) && (epoch < epochTo){
                totalRequests++
                if (elapsed <= slaTime) && (successStr == "true") {
                    slaAllowedRequests++
                }
                if successStr == "true" {
                  prc = append(prc, int(elapsed))
                }
            }
        }
    }
    if err := scanner.Err(); err != nil {
      log.Fatal(err)
    }
    prcCount := len(prc)
    
    // Вычисления перцентилей на основе, полученных из файла, и отфильтрованных по времени и по успешности транзакций, данных
    if (totalRequests > 0){
        resultPrc := (slaAllowedRequests * 100 / totalRequests)
        fmt.Printf("Процентиль, удовлетворяющий SLA: %d\n", resultPrc)
        fmt.Printf("Всего запросов: %d\nУдовлетворяют SLA (нет ошибок и время отклика меньше разрешенного): %d\n\n", totalRequests, slaAllowedRequests)
        
        sortedPrc := prc[:prcCount]
        sort.Ints(sortedPrc)
        result90Prc := int((float64(prcCount) * 90.0) / 100.0 + 0.5)
        if result90Prc > prcCount {result90Prc = prcCount + 1}
        fmt.Printf("Искомый процентиль (без учета запросов с ошибкой): %d мс\nЗапросов учтено (в рассчете процентиля): %d\n\n", sortedPrc[result90Prc - 1], result90Prc)
        
        if (showAllPercentiles){
            for i := 1;i <= 100; i++ {
                result90Prc := int((float64(prcCount) * float64(i)) / 100.0 + 0.5)
                if result90Prc > prcCount {result90Prc = prcCount + 1}
                fmt.Printf("%d й Искомый процентиль (без учета запросов с ошибкой):\tЗапросов учтено:%d\tПерцентиль:%d мс\n", i, result90Prc, sortedPrc[result90Prc - 1])
            }
        }
    } else {
      fmt.Println("В логе нет запросов, удовлетворяющих фильтру времени")
    }
}
