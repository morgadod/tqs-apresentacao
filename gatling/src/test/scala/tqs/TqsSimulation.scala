package tqs

import scala.concurrent.duration._

import scala.util.Random

import io.gatling.core.Predef._
import io.gatling.http.Predef._


class TqsSimulation extends Simulation {
    
    val httpProtocol = http
        .baseUrl("http://localhost:8000/api") 
        .acceptHeader("application/json")
    
    
    val votingScenario = scenario("Voting Scenario")
        .feed(tsv("votes.tsv").circular())
        .exec(
            http("Voting Request")
            .post("/survey")
            .header("content-type", "application/json")
            .body(StringBody("#{payload}"))
            .check(status.in(200, 201, 400, 404)) 
        )

    setUp(
        votingScenario.inject(
            constantUsersPerSec(2).during(10.seconds), 
            constantUsersPerSec(5).during(15.seconds),

            rampUsersPerSec(6).to(5000).during(3.minutes) 
        ).protocols(httpProtocol)
    )
}