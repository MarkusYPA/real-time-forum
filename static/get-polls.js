async function getPoll() {
    // Fetch stories from until enough of type poll have been found

    const maxIdResp = await fetch("https://hacker-news.firebaseio.com/v0/maxitem.json?")
    let startId = await maxIdResp.json();
    console.log(maxId)

     "https://hacker-news.firebaseio.com/v0/item/${startId}"

    postsContainer.innerHTML = ""
    const response = await fetch("https://hacker-news.firebaseio.com/v0/newstories.json");
    let storyIds = await response.json();

    console.log(storyIds.length, "stories to filter")

    let index = 0
    let filteredStories = []
    let step = storiesPerPage * 5
    while (filteredStories.length < storiesPerPage) {

        if (index < storyIds.length) {
            if (index + step >= storyIds.length) {
                step = storyIds.length - index
            }

            let newStories = await Promise.all(storyIds.slice(index, index + step).map(fetchStory))

            newStories.forEach((story) => {
                console.log(index, story.type)
                if (story.hasOwnProperty('type') && story.type == 'poll') {
                    filteredStories.push(story)
                }
            })
        } else {
            console.log("failed to find")
            break
        }

        index += step
        console.log(filteredStories)
    }
    console.log(filteredStories)
}

getPolls()