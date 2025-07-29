package computing

type ComputingRepo interface {

	//TODO 所有集群的总和显卡数量（分型号） "sum by (cluster) (node_memory_MemTotal_bytes)"
	//TODO 所有集群的总和显存（分型号） "sum by (cluster) (node_memory_MemTotal_bytes)"
	//TODO 所有集群的cpu core数量（分型号）
	//TODO 所有集群的内存 （分型号）

	//TODO GPU使用（总量）
	//TODO 显存使用（总量）
	//TODO CPU使用（总量）
	//TODO 内存使用（总量）

	//TODO 集群node数（按架构）
	//TODO 集群node数（按集群）

	//TODO 按集群CPU使用率
	//TODO 按集群GPU使用率
	//TODO 按集群显存使用率
	//TODO 按集群内存使用率

	//TODO 异常节点
	//TODO 异常算力卡

	//TODO 获取workspace
	//TODO 获取资源池

}
