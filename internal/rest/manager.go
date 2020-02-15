package rest

import "github.com/gin-gonic/gin"

// ExecutePipeline - executes a pipeline
// @Summary executes a pipeline
// @Description This method will execute the pipeline sent via the post body
// @ID execute-pipeline
// @Accept plain
// @Produce json
// @Success 200
// @Router /pipeline/execute [post]
func ExecutePipeline(c *gin.Context) {
}

// CreateManagerAPI - adds the manager api to a gin engine
func CreateManagerAPI(r *gin.Engine) {
	r.POST("/manager/pipeline/execute", ExecutePipeline)
}
