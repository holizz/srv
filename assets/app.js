var directory = angular.module('directory', [])

directory.controller('DirectoryListing', function ($scope, $http) {
    $scope.path = document.location.pathname

    $scope.files = []

    $http.get('/_srv/api?path='+encodeURI($scope.path))
        .success(function (data) {
            $scope.files = data
        })
        .error(function () {
            alert("oh no")
        })
})
