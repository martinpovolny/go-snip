# design decisions

in reality these should follow the project/team/company standards

1. use Gorm for DB/orm
2. use Viper for config
3. start with sqlite
4. use default autoincrement PK
5. simplification for sqlite: uuid = string
6. use the same struct for db dto and json representation
7. use only basic logger (no structured logging)

# interpretation of instructions

in reality these would need clarifications

 * in GET /{id}, id maps to external_id
 * unique constraint on external_id, not null for other fields
 * 201 and no data returned when record created
 * testing "program as close to how it will run in production as possible":
   * interpretation: actually run the program in a container and test externally w/ a http client

# test coverage

 * 1 example unit test

integration tests

 * POST
   * missing data -- done
   * invalid date -- done
   * invalid uuid -- n/a (we use string as external id)
   * uuid exists -- done
   * wrong method -- done

* GET
   * missing id -- done
   * invalid id -- done
   * wrong method -- done
   * json data match -- done

# TODO
   * metrics instrumentation