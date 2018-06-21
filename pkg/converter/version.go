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
	// TranslationVersion is a seed for Translation names and
	// backend resource IDs.
	// Changing TranslationVersion changes every Translation names
	// and backend resource IDs this controller would produce.  That is,
	// it effectively deletes every Translations and their backend resources
	// and create them with different IDs.  While it would cause severe
	// user traffic interruptions for a while, it can be useful when
	// upgrading the controller with incompatible Translations.
	TranslationVersion = "3"
)
