/*  logger.go
*
* @Author:             Nanang Suryadi
* @Date:               November 21, 2019
* @Last Modified by:   @suryakencana007
* @Last Modified time: 21/11/19 07:30
 */

package suki

type Logging interface {
        With(fields ...interface{}) Logging
        Debug(msg string, fields ...interface{})
        Info(msg string, fields ...interface{})
        Warn(msg string, fields ...interface{})
        Error(msg string, fields ...interface{})
        Fatal(msg string, fields ...interface{})
        Panic(msg string, fields ...interface{})
        Field(key string, value interface{}) interface{} // for type fields
}

var (
        logger *zapLog
)

func Instance() *zapLog {
        return NewZap(ProductionCore())
}

func With(fields ...interface{}) Logging {
        return Instance().With(fields...)
}

func Field(key string, value interface{}) interface{} {
        return Instance().Field(key, value)
}

func Debug(msg string, fields ...interface{}) {
        Instance().Debug(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
        Instance().Info(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
        Instance().Warn(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
        Instance().Error(msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
        Instance().Fatal(msg, fields...)
}

func Panic(msg string, fields ...interface{}) {
        Instance().Panic(msg, fields...)
}
