/*
Copyright 2026 The TabTabAI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8sadapter

const (
	STS_NAME_PREFIX = "" // NOTE: do not edit

	CONTAINER_NAME = "tabclaw"

	TAB_CLAW              = "tabtabai.com/tabclaw"
	TAB_CLAW_NAME         = "tabtabai.com/tabclaw-name"
	TAB_CLAW_OCCUPIED     = "tabtabai.com/tabclaw-occupied"
	TAB_CLAW_INIT_TRIGGER = "tabtabai.com/tabclaw-init"
	TAB_CLAW_ALLOC_STATUS = "tabtabai.com/tabclaw-alloc-status"

	AllocStatusAllocating = "allocating"
	AllocStatusAllocated  = "allocated"
	AllocStatusIdle       = "idle"

	ScheduledDeletionTime        = "tabtab.app.scheduled.deletion.time"
	ScheduledDeletionTimeTrigger = "tabtab.app.scheduled.deletion"

	CLAW_VOLUME_PREFIX = "claw:volume"
	CLAW_FINALIZERS    = "tabtabai.com/auto-remove-pv"

	// AnnotationGatewayURL is the annotation key on a StatefulSet that holds
	// the HTTPS WebUI access URL set by the operator.
	AnnotationGatewayURL = "tabtabai.com/gateway-url"
	// AnnotationWssURL is the annotation key on a StatefulSet that holds
	// the WebSocket URL set by the operator.
	AnnotationWssURL = "tabtabai.com/wss-url"
)

var NotAllowedStatefulsetNames = map[string]struct{}{
	"www":     {},
	"staging": {},
}
