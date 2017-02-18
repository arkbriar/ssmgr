'use strict';

angular.module('user').
factory('User', ['$resource',
  ($resource) => {
    return $resource('user/:userId', {}, {
      query: {
        method: 'POST',
        params: {userId: 'users'},
        isArray: true
      }
    });
  }
]);