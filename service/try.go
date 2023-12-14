package service

// TryCatchFinally 处理不可遇见性的错误时尝试使用，正常情况通过预见的错误和业务逻辑中的问题，建议使用 error 类型
// 谨慎使用panic defer/*
type TryCatchFinally struct {
	try     func()
	catch   func(any2 any)
	finally func()
}

func Try(try func()) *TryCatchFinally {
	tcf := &TryCatchFinally{try: try}
	return tcf
}

func (tcf *TryCatchFinally) Catch(catch func(interface{})) *TryCatchFinally {
	tcf.catch = catch
	return tcf
}

func (tcf *TryCatchFinally) Finally(finally func()) *TryCatchFinally {
	tcf.finally = finally
	return tcf
}

func (tcf *TryCatchFinally) Run() {
	defer func() {
		if r := recover(); r != nil {
			if tcf.catch != nil {
				tcf.catch(r)
			}
		}

		if tcf.finally != nil {
			tcf.finally()
		}
	}()

	tcf.try()
}
