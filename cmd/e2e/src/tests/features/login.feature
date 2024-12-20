# Copyright 2024 Specter Ops, Inc.
#
# Licensed under the Apache License, Version 2.0
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0

Feature: BHCE Login Page Tests

  Background:
    Given Create a new user with "Administrator" role 
    And User navigates to the login page

  @e2e  
  Scenario: User Authenticate with valid credentials
    And User enters valid username
    And User enters valid password
    When User clicks on the login button
    Then User is redirect to "explore" page

  @e2e 
  Scenario: User Authenticates with invalid credentials
    And User enters valid username
    And User enters invalid password
    When User clicks on the login button
    Then Page Displays Error Message
    
  @e2e 
  Scenario: Unauthenticated user accessing pages that require authentication
    When User visits "my-profile" page
    Then User is redirect back to "login" page
  @e2e 
  Scenario: User Authenticate with valid credentials
    And User enters valid username
    And User enters valid password
    When User clicks on the login button
    When User reloads the "explore" page
    Then User should be logged
