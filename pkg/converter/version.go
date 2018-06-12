// Copyright (C) 2018 Midokura SARL.
// All rights reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package converter

const (
	// Changing TranslationVersion would change every Translation names
	// and backend resource IDs.  That is, it effectively deletes and creates
	// every Translations and their backend resources.  While it would
	// cause user traffic interruptions, it can be useful when upgrading
	// the controller with incompatible Translations.
	TranslationVersion = "1"
)
