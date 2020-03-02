package logic

import (
	"time"

	"github.com/aleksaan/statusek/models"
	rc "github.com/aleksaan/statusek/returncodes"
)

func checkStatusIsBelongsToInstance(instanceInfo *models.InstanceInfo, statusInfo *models.StatusInfo) (bool, rc.ReturnCode) {
	if instanceInfo.Instance.ObjectID == statusInfo.Status.ObjectID {
		return true, rc.STATUS_IS_ACCORDING_TO_INSTANCE
	}
	return false, rc.STATUS_IS_NOT_ACCORDING_TO_INSTANCE
}

// CheckInstanceIsFinished - checks if instance finished or not
// Finished is if all of mandatory statuses of last level is set or if no one mandatory
// then at least one of optional statuses is set

func checkInstanceIsFinished(instanceInfo *models.InstanceInfo) (bool, rc.ReturnCode) {

	chk, rc0 := checkInstanceIsNotTimeout(instanceInfo)
	if chk == false {
		return true, rc0
	}

	var s = &models.StatusInfo{}
	//getting last statuses
	db.Raw("SELECT * FROM statuses.v_last_statuses WHERE object_id = ?", instanceInfo.Instance.ObjectID).Scan(&s.PrevStatuses)

	//checking previos statuses
	chkP, _ := checkPreviosStatusesIsSet(instanceInfo, s)

	if chkP == true {
		return true, rc.INSTANCE_IS_FINISHED
	}

	return false, rc.INSTANCE_IS_NOT_FINISHED

}

func checkInstanceIsNotTimeout(instanceInfo *models.InstanceInfo) (bool, rc.ReturnCode) {
	t1 := time.Now()
	t2 := *instanceInfo.Instance.InstanceCreationDt
	//fmt.Printf("\n\nTime 1: %s\nTime 2: %s\n\n", t1.Format(time.RFC3339), t2.Format(time.RFC3339))
	diff := t1.Sub(t2).Seconds()
	if diff < float64(instanceInfo.Instance.InstanceTimeout) {
		return true, rc.SUCCESS
	}

	return false, rc.INSTANCE_IS_IN_TIMEOUT
}

func checkPreviosStatusesIsSet(instanceInfo *models.InstanceInfo, statusInfo *models.StatusInfo) (bool, rc.ReturnCode) {
	var countPrevMandatory int
	var countPrevMandatoryIsSet int
	var countPrevOptional int
	var countPrevOptionalIsSet int
	for _, s := range statusInfo.PrevStatuses {
		if s.StatusIsMandatory {
			countPrevMandatory++
			for _, e := range instanceInfo.Events {
				if e.StatusID == s.StatusID {
					countPrevMandatoryIsSet++
					break
				}
			}
		} else {
			countPrevOptional++
			for _, e := range instanceInfo.Events {
				if e.StatusID == s.StatusID {
					countPrevOptionalIsSet++
					break
				}
			}
		}
	}

	if countPrevMandatory > countPrevMandatoryIsSet {
		return false, rc.NOT_ALL_PREVIOS_MANDATORY_STATUSES_IS_SET
		//"Не все обязательные статусы предыдущего уровня установлены"
	}

	if (countPrevMandatory == 0) && (countPrevOptional > 0) && (countPrevOptionalIsSet == 0) {
		return false, rc.NO_ONE_PREVIOS_OPTIONAL_STATUSES_IS_SET
		//"Не установлен ни один опциональный статус предыдущего уровня"
	}

	return true, rc.ALL_PREVIOS_STATUSES_IS_SET
}

func checkNextStatusesIsNotSet(instanceInfo *models.InstanceInfo, statusInfo *models.StatusInfo) (bool, rc.ReturnCode) {

	if statusInfo.Status.StatusIsMandatory {
		return true, rc.NEXT_STATUSES_IS_NOT_SET
	}

	for _, s := range statusInfo.NextStatuses {
		for _, e := range instanceInfo.Events {
			if e.StatusID == s.StatusID {
				return false, rc.AT_LEAST_ONE_NEXT_STATUS_IS_SET
			}
		}
	}

	return true, rc.NEXT_STATUSES_IS_NOT_SET
}

func checkCurrentStatusIsNotSet(instanceInfo *models.InstanceInfo, statusInfo *models.StatusInfo) (bool, rc.ReturnCode) {

	for _, e := range instanceInfo.Events {
		if e.StatusID == statusInfo.Status.StatusID {
			return false, rc.CURRENT_STATUS_IS_SET
		}
	}

	return true, rc.CURRENT_STATUS_IS_NOT_SET
}
