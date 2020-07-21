Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining

  Scenario: Eat 10 out of 15
    Given there are 15 godogs
    When I eat 10
    And I buy 5 more
    Then there should be 10 remaining
