// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

Cypress.Commands.add('apiInitSetup', ({
    loginAfter = false,
    userPrefix = 'user',
    teamPrefix = {name: 'team', displayName: 'Team'},
    channelPrefix = {name: 'channel', displayName: 'Channel'},
    createIncident = false,
    incidentDesc = '',
} = {}) => {
    return cy.apiCreateTeam(teamPrefix.name, teamPrefix.displayName).then(({team}) => {
        // # Add public channel
        return cy.apiCreateChannel(team.id, channelPrefix.name, channelPrefix.displayName).then(({channel}) => {
            return cy.apiCreateUser({prefix: userPrefix}).then(({user}) => {
                return cy.apiAddUserToTeam(team.id, user.id).then(() => {
                    return cy.apiAddUserToChannel(channel.id, user.id).then(() => {
                        return cy.apiCreateTestPlaybook(team.id, user.id).then(({playbook}) => {
                            if (createIncident) {
                                return cy.apiLogin(user).then(() => {
                                    return cy.apiCreateIncident(team.id, user.id, playbook.id, incidentDesc).then(({incident}) => {
                                            return cy.wrap({team, channel, user, playbook, incident});
                                    });
                                });
                            }
                            return cy.wrap({team, channel, user, playbook});
                        });
                        // if (loginAfter) {
                        //     return cy.apiLogin(user).then(() => {
                        //         return cy.wrap({team, channel, user});
                        //     });
                        // }
                        return cy.wrap({team, channel, user});
                    });
                });
            });
        });
    });
});
